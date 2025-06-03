import 'package:flutter/material.dart';
import 'package:owlistic/models/zettelkasten.dart';

/// Side panel for Zettelkasten graph controls and information
class ZettelkastenSidePanel extends StatefulWidget {
  final ZettelNode? selectedNode;
  final List<ZettelTag> availableTags;
  final String selectedFilter;
  final List<String> selectedTags;
  final Function(String) onFilterChanged;
  final Function(List<String>) onTagsChanged;
  final Function(ZettelNode?) onNodeSelected;
  final Function(String) onSearchChanged;
  final VoidCallback onClose;

  const ZettelkastenSidePanel({
    super.key,
    this.selectedNode,
    required this.availableTags,
    required this.selectedFilter,
    required this.selectedTags,
    required this.onFilterChanged,
    required this.onTagsChanged,
    required this.onNodeSelected,
    required this.onSearchChanged,
    required this.onClose,
  });

  @override
  State<ZettelkastenSidePanel> createState() => _ZettelkastenSidePanelState();
}

class _ZettelkastenSidePanelState extends State<ZettelkastenSidePanel> {
  final TextEditingController _searchController = TextEditingController();
  bool _showFilters = true;
  bool _showNodeInfo = true;
  bool _showConnections = true;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border(
          left: BorderSide(
            color: Theme.of(context).dividerColor,
            width: 1,
          ),
        ),
      ),
      child: Column(
        children: [
          // Header
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceVariant,
              border: Border(
                bottom: BorderSide(
                  color: Theme.of(context).dividerColor,
                  width: 1,
                ),
              ),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.account_tree,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Graph Controls',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: widget.onClose,
                  iconSize: 20,
                  tooltip: 'Close panel',
                ),
              ],
            ),
          ),

          // Content
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Search
                  TextField(
                    controller: _searchController,
                    decoration: const InputDecoration(
                      hintText: 'Search nodes...',
                      prefixIcon: Icon(Icons.search),
                      border: OutlineInputBorder(),
                      isDense: true,
                    ),
                    onChanged: (value) {
                      widget.onSearchChanged(value);
                    },
                  ),
                  
                  const SizedBox(height: 24),

                  // Filters Section
                  _buildCollapsibleSection(
                    title: 'Filters',
                    isExpanded: _showFilters,
                    onToggle: () => setState(() => _showFilters = !_showFilters),
                    child: _buildFiltersContent(),
                  ),

                  const SizedBox(height: 16),

                  // Node Information Section
                  if (widget.selectedNode != null) ...[
                    _buildCollapsibleSection(
                      title: 'Node Information',
                      isExpanded: _showNodeInfo,
                      onToggle: () => setState(() => _showNodeInfo = !_showNodeInfo),
                      child: _buildNodeInfoContent(),
                    ),
                    
                    const SizedBox(height: 16),

                    // Connections Section
                    _buildCollapsibleSection(
                      title: 'Connections',
                      isExpanded: _showConnections,
                      onToggle: () => setState(() => _showConnections = !_showConnections),
                      child: _buildConnectionsContent(),
                    ),
                  ] else ...[
                    // No node selected message
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.surfaceVariant,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Column(
                        children: [
                          Icon(
                            Icons.info_outline,
                            color: Theme.of(context).colorScheme.onSurfaceVariant,
                            size: 48,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Select a node to view details and connections',
                            textAlign: TextAlign.center,
                            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                              color: Theme.of(context).colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCollapsibleSection({
    required String title,
    required bool isExpanded,
    required VoidCallback onToggle,
    required Widget child,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          onTap: onToggle,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              children: [
                Icon(
                  isExpanded ? Icons.expand_less : Icons.expand_more,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ),
        if (isExpanded) child,
      ],
    );
  }

  Widget _buildFiltersContent() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Node Type Filter
        Text(
          'Node Type',
          style: Theme.of(context).textTheme.labelMedium,
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          children: [
            _buildFilterChip('all', 'All'),
            _buildFilterChip('note', 'Notes'),
            _buildFilterChip('task', 'Tasks'),
            _buildFilterChip('project', 'Projects'),
          ],
        ),
        
        const SizedBox(height: 16),

        // Tag Filter
        Text(
          'Tags',
          style: Theme.of(context).textTheme.labelMedium,
        ),
        const SizedBox(height: 8),
        if (widget.availableTags.isEmpty) ...[
          Text(
            'No tags available',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ),
        ] else ...[
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: widget.availableTags.map((tag) {
              final isSelected = widget.selectedTags.contains(tag.name);
              return FilterChip(
                label: Text(tag.name),
                selected: isSelected,
                onSelected: (selected) {
                  final newTags = List<String>.from(widget.selectedTags);
                  if (selected) {
                    newTags.add(tag.name);
                  } else {
                    newTags.remove(tag.name);
                  }
                  widget.onTagsChanged(newTags);
                },
                backgroundColor: Color(int.parse(tag.color.replaceFirst('#', '0xFF'))),
                checkmarkColor: Colors.white,
              );
            }).toList(),
          ),
        ],

        const SizedBox(height: 16),

        // Clear Filters Button
        if (widget.selectedFilter != 'all' || widget.selectedTags.isNotEmpty)
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: () {
                widget.onFilterChanged('all');
                widget.onTagsChanged([]);
              },
              child: const Text('Clear Filters'),
            ),
          ),
      ],
    );
  }

  Widget _buildFilterChip(String value, String label) {
    final isSelected = widget.selectedFilter == value;
    return ChoiceChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          widget.onFilterChanged(value);
        }
      },
    );
  }

  Widget _buildNodeInfoContent() {
    final node = widget.selectedNode!;
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Node Title
        Text(
          node.title,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        
        const SizedBox(height: 8),

        // Node Type
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          decoration: BoxDecoration(
            color: _getNodeTypeColor(node.nodeType).withOpacity(0.2),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            NodeType.getDisplayName(node.nodeType),
            style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: _getNodeTypeColor(node.nodeType),
              fontWeight: FontWeight.bold,
            ),
          ),
        ),

        const SizedBox(height: 12),

        // Summary
        if (node.summary != null && node.summary!.isNotEmpty) ...[
          Text(
            'Summary',
            style: Theme.of(context).textTheme.labelMedium,
          ),
          const SizedBox(height: 4),
          Text(
            node.summary!,
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: 12),
        ],

        // Tags
        if (node.tags.isNotEmpty) ...[
          Text(
            'Tags',
            style: Theme.of(context).textTheme.labelMedium,
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: node.tags.map((tag) {
              return Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Color(int.parse(tag.color.replaceFirst('#', '0xFF'))),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  tag.name,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                  ),
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
        ],

        // Metadata
        Text(
          'Created',
          style: Theme.of(context).textTheme.labelMedium,
        ),
        const SizedBox(height: 4),
        Text(
          _formatDate(node.createdAt),
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }

  Widget _buildConnectionsContent() {
    final node = widget.selectedNode!;
    final outgoingConnections = node.connections;
    final incomingConnections = node.backLinks;
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Outgoing Connections
        if (outgoingConnections.isNotEmpty) ...[
          Text(
            'Outgoing (${outgoingConnections.length})',
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          ...outgoingConnections.map((connection) => 
            _buildConnectionItem(connection, isOutgoing: true)
          ),
          const SizedBox(height: 16),
        ],

        // Incoming Connections
        if (incomingConnections.isNotEmpty) ...[
          Text(
            'Incoming (${incomingConnections.length})',
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          ...incomingConnections.map((connection) => 
            _buildConnectionItem(connection, isOutgoing: false)
          ),
          const SizedBox(height: 16),
        ],

        // No connections message
        if (outgoingConnections.isEmpty && incomingConnections.isEmpty) ...[
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceVariant,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.link_off,
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'No connections yet',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],

        // Add Connection Button
        const SizedBox(height: 16),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            onPressed: () {
              // TODO: Show connection creation dialog
              _showCreateConnectionDialog();
            },
            icon: const Icon(Icons.add_link),
            label: const Text('Add Connection'),
          ),
        ),
      ],
    );
  }

  Widget _buildConnectionItem(ZettelEdge connection, {required bool isOutgoing}) {
    final targetNode = isOutgoing ? connection.targetNode : connection.sourceNode;
    if (targetNode == null) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        border: Border.all(color: Theme.of(context).dividerColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                _getConnectionIcon(connection.connectionType),
                size: 16,
                color: _getConnectionTypeColor(connection.connectionType),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  targetNode.title,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w500,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: _getConnectionTypeColor(connection.connectionType).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  ConnectionType.getDisplayName(connection.connectionType),
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: _getConnectionTypeColor(connection.connectionType),
                  ),
                ),
              ),
              const Spacer(),
              if (connection.strength < 1.0)
                Text(
                  '${(connection.strength * 100).toInt()}%',
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
            ],
          ),
          if (connection.description != null && connection.description!.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              connection.description!,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ],
      ),
    );
  }

  Color _getNodeTypeColor(String nodeType) {
    switch (nodeType) {
      case 'note':
        return Colors.blue;
      case 'task':
        return Colors.green;
      case 'project':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  Color _getConnectionTypeColor(String connectionType) {
    switch (connectionType) {
      case 'related':
        return Colors.grey;
      case 'depends_on':
        return Colors.red;
      case 'references':
        return Colors.blue;
      case 'supports':
        return Colors.green;
      case 'contradicts':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  IconData _getConnectionIcon(String connectionType) {
    switch (connectionType) {
      case 'related':
        return Icons.link;
      case 'depends_on':
        return Icons.arrow_forward;
      case 'references':
        return Icons.bookmark;
      case 'supports':
        return Icons.thumb_up;
      case 'contradicts':
        return Icons.block;
      default:
        return Icons.link;
    }
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }

  void _showCreateConnectionDialog() {
    // TODO: Implement connection creation dialog
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Create Connection'),
        content: const Text('Connection creation dialog will be implemented here.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Create'),
          ),
        ],
      ),
    );
  }
}