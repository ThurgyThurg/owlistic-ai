import 'dart:math';

import 'package:flutter/material.dart';
import 'package:follow_the_leader/follow_the_leader.dart';
import 'package:ionicons/ionicons.dart';
import 'package:overlord/follow_the_leader.dart';
import 'package:super_editor/super_editor.dart';
import 'package:owlistic/utils/super_editor_item_selector.dart';

/// Small toolbar that is intended to display near some selected
/// text and offer a few text formatting controls.
///
/// [EditorToolbar] expects to be displayed in a [Stack] where it
/// will position itself based on the given [anchor]. This can be
/// accomplished, for example, by adding [EditorToolbar] to the
/// application [Overlay]. Any other [Stack] should work, too.
class EditorToolbar extends StatefulWidget {
  const EditorToolbar({
    super.key,
    required this.editorViewportKey,
    required this.editorFocusNode,
    required this.editor,
    required this.document,
    required this.composer,
    required this.anchor,
    required this.closeToolbar,
  });

  /// [GlobalKey] that should be attached to a widget that wraps the viewport
  /// area, which keeps the toolbar from appearing outside of the editor area.
  final GlobalKey editorViewportKey;

  /// A [LeaderLink] that should be attached to the boundary of the toolbar
  /// focal area, such as wrapped around the user's selection area.
  ///
  /// The toolbar is positioned relative to this anchor link.
  final LeaderLink anchor;

  /// The [FocusNode] attached to the editor to which this toolbar applies.
  final FocusNode editorFocusNode;

  /// The [editor] is used to alter document content, such as
  /// when the user selects a different block format for a
  /// text blob, e.g., paragraph, header, blockquote, or
  /// to apply styles to text.
  final Editor? editor;

  final Document document;

  /// The [composer] provides access to the user's current
  /// selection within the document, which dictates the
  /// content that is altered by the toolbar's options.
  final DocumentComposer composer;

  /// Delegate that instructs the owner of this [EditorToolbar]
  /// to close the toolbar, such as after submitting a URL
  /// for some text.
  final VoidCallback closeToolbar;

  @override
  State<EditorToolbar> createState() => _EditorToolbarState();
}

class _EditorToolbarState extends State<EditorToolbar> {
  late final FollowerAligner _toolbarAligner;
  late FollowerBoundary _screenBoundary;

  bool _showUrlField = false;
  late FocusNode _popoverFocusNode;
  late FocusNode _urlFocusNode;
  ImeAttributedTextEditingController? _urlController;

  @override
  void initState() {
    super.initState();

    _toolbarAligner = CupertinoPopoverToolbarAligner(widget.editorViewportKey);

    _popoverFocusNode = FocusNode();

    _urlFocusNode = FocusNode();
    _urlController =
        ImeAttributedTextEditingController(controller: SingleLineAttributedTextEditingController(_applyLink)) //
          ..onPerformActionPressed = _onPerformAction
          ..text = AttributedText("https://");
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    _screenBoundary = WidgetFollowerBoundary(
      boundaryKey: widget.editorViewportKey,
      devicePixelRatio: MediaQuery.devicePixelRatioOf(context),
    );
  }

  @override
  void dispose() {
    _urlFocusNode.dispose();
    _urlController?.dispose();
    _popoverFocusNode.dispose();

    super.dispose();
  }

  /// Returns true if the currently selected text node is capable of being
  /// transformed into a different type text node, returns false if
  /// multiple nodes are selected, no node is selected, or the selected
  /// node is not a standard text block.
  bool _isConvertibleNode() {
    final selection = widget.composer.selection;
    if (selection == null || selection.base.nodeId != selection.extent.nodeId) {
      return false;
    }

    final selectedNode = widget.document.getNodeById(selection.extent.nodeId);
    return (
      selectedNode is ParagraphNode ||
      selectedNode is ListItemNode ||
      selectedNode is TaskNode
    );
  }

  /// Returns the block type of the currently selected text node.
  ///
  /// Throws an exception if the currently selected node is not a text node.
  _TextType _getCurrentTextType() {
    final selection = widget.composer.selection;
    if (selection == null) {
      return _TextType.paragraph;
    }
    final selectedNode = widget.document.getNodeById(selection.extent.nodeId);
    if (selectedNode is ParagraphNode) {
      final type = selectedNode.getMetadataValue('blockType');

      if (type == header1Attribution) {
        return _TextType.header1;
      } else if (type == header2Attribution) {
        return _TextType.header2;
      } else if (type == header3Attribution) {
        return _TextType.header3;
      } else if (type == blockquoteAttribution) {
        return _TextType.blockquote;
      } else {
        return _TextType.paragraph;
      }
    } else if (selectedNode is ListItemNode) {
      return selectedNode.type == ListItemType.ordered ? _TextType.orderedListItem : _TextType.unorderedListItem;
    } else if (selectedNode is TaskNode) {
      return _TextType.taskItem;
    } else {
      throw Exception('Alignment does not apply to node of type: $selectedNode');
    }
  }

  /// Returns the text alignment of the currently selected text node.
  /// Throws an exception if the currently selected node is not a text node.
  TextAlign _getCurrentTextAlignment() {
    final selection = widget.composer.selection;
    if (selection == null) {
      return TextAlign.left;
    }
    final selectedNode = widget.document.getNodeById(selection.extent.nodeId);
    if (selectedNode is ParagraphNode) {
      final align = selectedNode.getMetadataValue('textAlign');
      switch (align) {
        case 'left':
          return TextAlign.left;
        case 'center':
          return TextAlign.center;
        case 'right':
          return TextAlign.right;
        case 'justify':
          return TextAlign.justify;
        default:
          return TextAlign.left;
      }
    } else {
      throw Exception('Invalid node type: $selectedNode');
    }
  }

  /// Returns true if a single text node is selected and that text node
  /// is capable of respecting alignment, returns false otherwise.
  bool _isTextAlignable() {
    final selection = widget.composer.selection;
    if (selection == null || selection.base.nodeId != selection.extent.nodeId) {
      return false;
    }

    final selectedNode = widget.document.getNodeById(selection.extent.nodeId);
    return selectedNode is ParagraphNode;
  }

  // Create a new node of [type] at the current caret position
  void _createNodeFromType(_TextType? type){
    if (type == null) return;

    dynamic node;

    switch (type){
      case _TextType.header1:
      case _TextType.header2:
      case _TextType.header3:
      case _TextType.blockquote:
        node = ParagraphNode(
          id: Editor.createNodeId(),
          text: AttributedText(''),
          metadata: {
            'blockType': _getBlockTypeAttribution(type),
          }
        );
        break;
      case _TextType.orderedListItem:
      case _TextType.unorderedListItem:
        node = ListItemNode(
          id: Editor.createNodeId(),
          text: AttributedText(''),
          itemType: type == _TextType.orderedListItem ? ListItemType.ordered : ListItemType.unordered,
        );
        break;
      case _TextType.taskItem:
       node = TaskNode(
          id: Editor.createNodeId(),
          text: AttributedText(''),
          isComplete: false,
       );
       break;
      case _TextType.paragraph:
      default:
        node = ParagraphNode(
          id: Editor.createNodeId(),
          text: AttributedText('')
        );
    }
    widget.editor!.execute([
      InsertNodeAtCaretRequest(node: node),
    ]);
  }

  /// Converts the currently selected text node into a new type of
  /// text node, represented by [newType].
  ///
  /// For example: convert a paragraph to a blockquote, or a header
  /// to a list item.
  void _convertTextToNewType(_TextType? newType) {
    final selection = widget.composer.selection;
    if (selection == null || widget.editor == null) {
      return;
    }
    
    final existingTextType = _getCurrentTextType();

    if (existingTextType == newType) {
      // The text is already the desired type. Return.
      return;
    }

    if (_isListItem(existingTextType) && _isListItem(newType)) {
      widget.editor!.execute([
        ChangeListItemTypeRequest(
          nodeId: selection.extent.nodeId,
          newType: newType == _TextType.orderedListItem ? ListItemType.ordered : ListItemType.unordered,
        ),
      ]);
    } else if (_isListItem(existingTextType) && !_isListItem(newType)) {
      widget.editor!.execute([
        ConvertListItemToParagraphRequest(
          nodeId: selection.extent.nodeId,
          paragraphMetadata: {
            'blockType': _getBlockTypeAttribution(newType),
          },
        ),
      ]);
    } else if (!_isListItem(existingTextType) && _isListItem(newType)) {
      if (_isTaskItem(existingTextType)) {
        // TODO: this is a workaround as no ConvertTaskToListItemRequest exists
        widget.editor!.execute([
          ConvertTaskToParagraphRequest(
            nodeId: selection.extent.nodeId,
          ),
        ]);
      }
      widget.editor!.execute([
        ConvertParagraphToListItemRequest(
          nodeId: selection.extent.nodeId,
          type: newType == _TextType.orderedListItem ? ListItemType.ordered : ListItemType.unordered,
        ),
      ]);
    } else if (!_isTaskItem(existingTextType) && _isTaskItem(newType)) {
      widget.editor!.execute([
        ConvertParagraphToTaskRequest(
          nodeId: selection.extent.nodeId,
          isComplete: false,
        ),
      ]);
    } else if (_isTaskItem(existingTextType) && !_isTaskItem(newType)) {
      widget.editor!.execute([
        ConvertTaskToParagraphRequest(
          nodeId: selection.extent.nodeId,
          paragraphMetadata: {
            'blockType': _getBlockTypeAttribution(newType),
          }
        ),
      ]);
    } else {
      // Apply a new block type to an existing paragraph node.
      widget.editor!.execute([
        ChangeParagraphBlockTypeRequest(
          nodeId: selection.extent.nodeId,
          blockType: _getBlockTypeAttribution(newType),
        ),
      ]);
    }
  }

  /// Returns true if the given [_TextType] represents a
  /// task item, returns false otherwise.
  bool _isTaskItem(_TextType? type) {
    return type == _TextType.taskItem;
  }
  
  /// Returns true if the given [_TextType] represents an
  /// ordered or unordered list item, returns false otherwise.
  bool _isListItem(_TextType? type) {
    return type == _TextType.orderedListItem || type == _TextType.unorderedListItem;
  }

  /// Returns the text [Attribution] associated with the given
  /// [_TextType], e.g., [_TextType.header1] -> [header1Attribution].
  Attribution? _getBlockTypeAttribution(_TextType? newType) {
    switch (newType) {
      case _TextType.header1:
        return header1Attribution;
      case _TextType.header2:
        return header2Attribution;
      case _TextType.header3:
        return header3Attribution;
      case _TextType.blockquote:
        return blockquoteAttribution;
      case _TextType.paragraph:
      default:
        return null;
    }
  }

  /// Toggles bold styling for the current selected text.
  void _toggleBold() {
    final selection = widget.composer.selection;
    if (selection == null || widget.editor == null) {
      return;
    }
    
    widget.editor!.execute([
      ToggleTextAttributionsRequest(
        documentRange: selection,
        attributions: {boldAttribution},
      ),
    ]);
  }

  /// Toggles italic styling for the current selected text.
  void _toggleItalics() {
    final selection = widget.composer.selection;
    if (selection == null || widget.editor == null) {
      return;
    }
    
    widget.editor!.execute([
      ToggleTextAttributionsRequest(
        documentRange: selection,
        attributions: {italicsAttribution},
      ),
    ]);
  }

  /// Toggles strikethrough styling for the current selected text.
  void _toggleUnderline() {
    final selection = widget.composer.selection;
    if (selection == null || widget.editor == null) {
      return;
    }
    
    widget.editor!.execute([
      ToggleTextAttributionsRequest(
        documentRange: selection,
        attributions: {underlineAttribution},
      ),
    ]);
  }

  /// Toggles strikethrough styling for the current selected text.
  void _toggleStrikethrough() {
    final selection = widget.composer.selection;
    if (selection == null || widget.editor == null) {
      return;
    }
    
    widget.editor!.execute([
      ToggleTextAttributionsRequest(
        documentRange: selection,
        attributions: {strikethroughAttribution},
      ),
    ]);
  }

  /// Returns true if the current text selection includes part
  /// or all of a single link, returns false if zero links are
  /// in the selection or if 2+ links are in the selection.
  bool _isSingleLinkSelected() {
    return _getSelectedLinkSpans().length == 1;
  }

  /// Returns true if the current text selection includes 2+
  /// links, returns false otherwise.
  bool _areMultipleLinksSelected() {
    return _getSelectedLinkSpans().length >= 2;
  }

  /// Returns any link-based [AttributionSpan]s that appear partially
  /// or wholly within the current text selection.
  Set<AttributionSpan> _getSelectedLinkSpans() {
    final selection = widget.composer.selection!;
    final baseOffset = (selection.base.nodePosition as TextPosition).offset;
    final extentOffset = (selection.extent.nodePosition as TextPosition).offset;
    final selectionStart = min(baseOffset, extentOffset);
    final selectionEnd = max(baseOffset, extentOffset);
    final selectionRange = SpanRange(selectionStart, selectionEnd - 1);

    final textNode = widget.document.getNodeById(selection.extent.nodeId) as TextNode;
    final text = textNode.text;

    final overlappingLinkAttributions = text.getAttributionSpansInRange(
      attributionFilter: (Attribution attribution) => attribution is LinkAttribution,
      range: selectionRange,
    );

    return overlappingLinkAttributions;
  }

  /// Takes appropriate action when the toolbar's link button is
  /// pressed.
  void _onLinkPressed() {
    final selection = widget.composer.selection!;
    final baseOffset = (selection.base.nodePosition as TextPosition).offset;
    final extentOffset = (selection.extent.nodePosition as TextPosition).offset;
    final selectionStart = min(baseOffset, extentOffset);
    final selectionEnd = max(baseOffset, extentOffset);
    final selectionRange = SpanRange(selectionStart, selectionEnd - 1);

    final textNode = widget.document.getNodeById(selection.extent.nodeId) as TextNode;
    final text = textNode.text;

    final overlappingLinkAttributions = text.getAttributionSpansInRange(
      attributionFilter: (Attribution attribution) => attribution is LinkAttribution,
      range: selectionRange,
    );

    if (overlappingLinkAttributions.length >= 2) {
      // Do nothing when multiple links are selected.
      return;
    }

    if (overlappingLinkAttributions.isNotEmpty) {
      // The selected text contains one other link.
      final overlappingLinkSpan = overlappingLinkAttributions.first;
      final isLinkSelectionOnTrailingEdge =
          (overlappingLinkSpan.start >= selectionRange.start && overlappingLinkSpan.start <= selectionRange.end) ||
              (overlappingLinkSpan.end >= selectionRange.start && overlappingLinkSpan.end <= selectionRange.end);

      if (isLinkSelectionOnTrailingEdge) {
        // The selected text covers the beginning, or the end, or the entire
        // existing link. Remove the link attribution from the selected text.
        text.removeAttribution(overlappingLinkSpan.attribution, selectionRange);
      } else {
        // The selected text sits somewhere within the existing link. Remove
        // the entire link attribution.
        text.removeAttribution(
          overlappingLinkSpan.attribution,
          overlappingLinkSpan.range,
        );
      }
    } else {
      // There are no other links in the selection. Show the URL text field.
      setState(() {
        _showUrlField = true;
        _urlFocusNode.requestFocus();
      });
    }
  }

  /// Takes the text from the [urlController] and applies it as a link
  /// attribution to the currently selected text.
  void _applyLink() {
    final url = _urlController!.text.toPlainText();

    final selection = widget.composer.selection!;
    final baseOffset = (selection.base.nodePosition as TextPosition).offset;
    final extentOffset = (selection.extent.nodePosition as TextPosition).offset;
    final selectionStart = min(baseOffset, extentOffset);
    final selectionEnd = max(baseOffset, extentOffset);
    final selectionRange = TextRange(start: selectionStart, end: selectionEnd - 1);

    final textNode = widget.document.getNodeById(selection.extent.nodeId) as TextNode;
    final text = textNode.text;

    final trimmedRange = _trimTextRangeWhitespace(text, selectionRange);

    final linkAttribution = LinkAttribution(url);

    widget.editor!.execute([
      AddTextAttributionsRequest(
        documentRange: DocumentRange(
          start: DocumentPosition(
            nodeId: textNode.id,
            nodePosition: TextNodePosition(offset: trimmedRange.start),
          ),
          end: DocumentPosition(
            nodeId: textNode.id,
            nodePosition: TextNodePosition(offset: trimmedRange.end),
          ),
        ),
        attributions: {linkAttribution},
      ),
    ]);

    // Clear the field and hide the URL bar
    _urlController!.clearTextAndSelection();
    setState(() {
      _showUrlField = false;
      _urlFocusNode.unfocus(disposition: UnfocusDisposition.previouslyFocusedChild);
      widget.closeToolbar();
    });
  }

  /// Given [text] and a [range] within the [text], the [range] is
  /// shortened on both sides to remove any trailing whitespace and
  /// the new range is returned.
  SpanRange _trimTextRangeWhitespace(AttributedText text, TextRange range) {
    int startOffset = range.start;
    int endOffset = range.end;

    String rawText = text.toPlainText();

    while (startOffset < range.end && rawText[startOffset] == ' ') {
      startOffset += 1;
    }
    while (endOffset > startOffset && rawText[endOffset] == ' ') {
      endOffset -= 1;
    }

    // Add 1 to the end offset because SpanRange treats the end offset to be exclusive.
    return SpanRange(startOffset, endOffset + 1);
  }

  /// Changes the alignment of the current selected text node
  /// to reflect [newAlignment].
  void _changeAlignment(TextAlign? newAlignment) {
    final selection = widget.composer.selection;
    if (newAlignment == null || selection == null || widget.editor == null) {
      return;
    }

    widget.editor!.execute([
      ChangeParagraphAlignmentRequest(
        nodeId: selection.extent.nodeId,
        alignment: newAlignment,
      ),
    ]);
  }

  void _onPerformAction(TextInputAction action) {
    if (action == TextInputAction.done) {
      _applyLink();
    }
  }

  /// Called when the user selects a block type on the toolbar.
  void _onBlockTypeCreated(IconItem? selectedItem) {
    if (selectedItem != null) {
      setState(() {
        _createNodeFromType(_TextType.values //
            .where((e) => e.name == selectedItem.id)
            .first);
      });
    }
  }

  /// Called when the user selects a block type on the toolbar.
  void _onBlockTypeSelected(IconItem? selectedItem) {
    if (selectedItem != null) {
      setState(() {
        _convertTextToNewType(_TextType.values //
            .where((e) => e.name == selectedItem.id)
            .first);
      });
    }
  }

  /// Called when the user selects an alignment on the toolbar.
  void _onAlignmentSelected(IconItem? selectedItem) {
    if (selectedItem != null) {
      setState(() {
        _changeAlignment(TextAlign.values.firstWhere((e) => e.name == selectedItem.id));
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return BuildInOrder(
      children: [
        FollowerFadeOutBeyondBoundary(
          link: widget.anchor,
          boundary: _screenBoundary,
          child: Follower.withAligner(
            link: widget.anchor,
            aligner: _toolbarAligner,
            boundary: _screenBoundary,
            showWhenUnlinked: false,
            child: _buildToolbars(),
          ),
        ),
      ],
    );
  }

  Widget _buildToolbars() {
    return SuperEditorPopover(
      popoverFocusNode: _popoverFocusNode,
      editorFocusNode: widget.editorFocusNode,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildToolbar(),
          if (_showUrlField) ...[
            const SizedBox(height: 8),
            _buildUrlField(),
          ],
        ],
      ),
    );
  }

  Widget _buildToolbar() {
    return IntrinsicWidth(
      child: Material(
        shape: const StadiumBorder(),
        elevation: 5,
        clipBehavior: Clip.hardEdge,
        child: SizedBox(
          height: 40,
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Tooltip(
                    message: 'Add Block at cursor',
                    child: _buildNewBlockSelector(),
                  ),
                ],
              ),
              _buildVerticalDivider(),
              // Only allow the user to select a new type of text node if
              // the currently selected node can be converted.
              if (_isConvertibleNode()) ...[
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Tooltip(
                      message: 'Change Block type',
                      child: _buildBlockTypeSelector(),
                    )
                  ],
                ),
              ],
              _buildVerticalDivider(),
              // Text format options
              Center(
                child: IconButton(
                  onPressed: _toggleBold,
                  icon: const Icon(Icons.format_bold),
                  splashRadius: 16,
                  tooltip: 'Bold',
                ),
              ),
              Center(
                child: IconButton(
                  onPressed: _toggleItalics,
                  icon: const Icon(Icons.format_italic),
                  splashRadius: 16,
                  tooltip: 'Italics',
                ),
              ),
              Center(
                child: IconButton(
                  onPressed: _toggleUnderline,
                  icon: const Icon(Icons.format_underline_sharp),
                  splashRadius: 16,
                  tooltip: 'Underline',
                ),
              ),
              Center(
                child: IconButton(
                  onPressed: _toggleStrikethrough,
                  icon: const Icon(Icons.strikethrough_s),
                  splashRadius: 16,
                  tooltip: 'Strikethrough',
                ),
              ),
              _buildVerticalDivider(),
              // Only display alignment controls if the currently selected text
              // node respects alignment. List items, for example, do not.
              if (_isTextAlignable()) ...[
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Tooltip(
                      message: 'Text Alignment',
                      child: _buildAlignmentSelector(),
                    ),
                    _buildVerticalDivider(),
                  ],
                ),
              ],
              Center(
                child: IconButton(
                  onPressed: _areMultipleLinksSelected() ? null : _onLinkPressed,
                  icon: const Icon(Icons.link),
                  color: _isSingleLinkSelected() ? const Color(0xFF007AFF) : IconTheme.of(context).color,
                  splashRadius: 16,
                  tooltip: 'Link',
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAlignmentSelector() {
    final alignment = _getCurrentTextAlignment();
    return ItemSelector(
      parentFocusNode: widget.editorFocusNode,
      boundaryKey: widget.editorViewportKey,
      value: IconItem(
        id: alignment.name,
        icon: _buildTextAlignIcon(alignment),
      ),
      items: const [TextAlign.left, TextAlign.center, TextAlign.right, TextAlign.justify]
          .map(
            (alignment) => IconItem(
              icon: _buildTextAlignIcon(alignment),
              id: alignment.name,
            ),
          )
          .toList(),
      onSelected: _onAlignmentSelected,
    );
  }

  Widget _buildNewBlockSelector() {
    return ItemSelector(
      parentFocusNode: widget.editorFocusNode,
      boundaryKey: widget.editorViewportKey,
      value: const IconItem(
        id: "Add New Block",
        icon: Icons.add_circle,
      ),
      items: _TextType.values
          .map(
            (blockType) => IconItem(
              id: blockType.name,
              icon: _buildBlockTypeIcon(blockType),
            ),
          )
          .toList(),
      onSelected: _onBlockTypeCreated,
    );
  }

  Widget _buildBlockTypeSelector() {
    final currentBlockType = _getCurrentTextType();
    return ItemSelector(
      parentFocusNode: widget.editorFocusNode,
      boundaryKey: widget.editorViewportKey,
      value: IconItem(
        id: currentBlockType.name,
        icon: _buildBlockTypeIcon(currentBlockType),
      ),
      items: _TextType.values
          .map(
            (blockType) => IconItem(
              id: blockType.name,
              icon: _buildBlockTypeIcon(blockType),
            ),
          )
          .toList(),
      onSelected: _onBlockTypeSelected,
    );
  }

  Widget _buildUrlField() {
    return Material(
      shape: const StadiumBorder(),
      elevation: 5,
      clipBehavior: Clip.hardEdge,
      child: Container(
        width: 400,
        height: 40,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            Expanded(
              child: SuperTextField(
                focusNode: _urlFocusNode,
                textController: _urlController,
                minLines: 1,
                maxLines: 1,
                inputSource: TextInputSource.ime,
                hintBehavior: HintBehavior.displayHintUntilTextEntered,
                hintBuilder: (context) {
                  return const Text(
                    "enter a url...",
                    style: TextStyle(
                      color: Colors.grey,
                      fontSize: 16,
                    ),
                  );
                },
                textStyleBuilder: (_) {
                  return const TextStyle(
                    color: Colors.black,
                    fontSize: 16,
                  );
                },
              ),
            ),
            IconButton(
              icon: const Icon(Icons.close),
              iconSize: 20,
              splashRadius: 16,
              padding: EdgeInsets.zero,
              onPressed: () {
                setState(() {
                  _urlFocusNode.unfocus();
                  _showUrlField = false;
                  _urlController!.clearTextAndSelection();
                });
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerticalDivider() {
    return Container(
      width: 1,
      color: Colors.grey.shade300,
    );
  }

  IconData _buildTextAlignIcon(TextAlign align) {
    switch (align) {
      case TextAlign.left:
      case TextAlign.start:
        return Icons.format_align_left;
      case TextAlign.center:
        return Icons.format_align_center;
      case TextAlign.right:
      case TextAlign.end:
        return Icons.format_align_right;
      case TextAlign.justify:
        return Icons.format_align_justify;
    }
  }
}

IconData _buildBlockTypeIcon(_TextType type) {
  switch (type) {
    // TODO: find better H1 icon
    case _TextType.header1:
      return Icons.title;
    // TODO: find better H2 icon
    case _TextType.header2:
      return Icons.title;
    // TODO: find better H3 icon
    case _TextType.header3:
      return Icons.title;
    case _TextType.paragraph:
      return Ionicons.text;
    case _TextType.blockquote:
      return Icons.format_quote_rounded;
    case _TextType.orderedListItem:
      return Icons.format_list_numbered;
    case _TextType.unorderedListItem:
      return Icons.format_list_bulleted;
    case _TextType.taskItem:
      return Icons.checklist_outlined;
  }
}

enum _TextType {
  header1,
  header2,
  header3,
  paragraph,
  blockquote,
  orderedListItem,
  unorderedListItem,
  taskItem,
}

class SingleLineAttributedTextEditingController extends AttributedTextEditingController {
  SingleLineAttributedTextEditingController(this.onSubmit);

  final VoidCallback onSubmit;

  @override
  void insertNewline() {
    // Don't insert newline in a single-line text field.

    // Invoke callback to take action on enter.
    onSubmit();

    // TODO: this is a hack. SuperTextField shouldn't insert newlines in a single
    // line field (#697).
  }
}