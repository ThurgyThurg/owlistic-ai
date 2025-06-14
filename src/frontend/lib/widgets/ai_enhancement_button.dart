import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/ai_service.dart';
import '../utils/logger.dart';

class AIEnhancementButton extends StatefulWidget {
  final String noteId;
  final VoidCallback? onProcessingComplete;
  final VoidCallback? onBeforeProcessing;
  final Widget? icon;
  final String? tooltip;

  const AIEnhancementButton({
    Key? key,
    required this.noteId,
    this.onProcessingComplete,
    this.onBeforeProcessing,
    this.icon,
    this.tooltip,
  }) : super(key: key);

  @override
  State<AIEnhancementButton> createState() => _AIEnhancementButtonState();
}

class _AIEnhancementButtonState extends State<AIEnhancementButton>
    with SingleTickerProviderStateMixin {
  final Logger _logger = Logger('AIEnhancementButton');
  bool _isProcessing = false;
  late AnimationController _animationController;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    );
    _animation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
    ));
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  Future<void> _processWithAI() async {
    if (_isProcessing) return;

    // Call onBeforeProcessing to save any pending changes
    widget.onBeforeProcessing?.call();
    
    // Add a small delay to ensure save operations complete
    await Future.delayed(const Duration(milliseconds: 100));

    setState(() {
      _isProcessing = true;
    });

    _animationController.repeat();

    try {
      final result = await AIService().processNoteWithAI(widget.noteId);
      _logger.info('AI processing result: $result');
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Row(
              children: [
                Icon(Icons.auto_awesome, color: Colors.white),
                SizedBox(width: 8),
                Text('AI processing started! Check back in a moment.'),
              ],
            ),
            backgroundColor: Colors.green,
            duration: Duration(seconds: 3),
          ),
        );

        widget.onProcessingComplete?.call();
      }
    } catch (e) {
      _logger.error('AI processing failed: $e');
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(Icons.error, color: Colors.white),
                const SizedBox(width: 8),
                Expanded(
                  child: Text('AI processing failed: ${e.toString()}'),
                ),
              ],
            ),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 5),
            action: SnackBarAction(
              label: 'Retry',
              textColor: Colors.white,
              onPressed: _processWithAI,
            ),
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isProcessing = false;
        });
        _animationController.stop();
        _animationController.reset();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: widget.tooltip ?? 'Enhance with AI',
      child: AnimatedBuilder(
        animation: _animation,
        builder: (context, child) {
          return FloatingActionButton.small(
            onPressed: _isProcessing ? null : _processWithAI,
            backgroundColor: _isProcessing 
                ? Theme.of(context).colorScheme.surface
                : Theme.of(context).colorScheme.primary,
            foregroundColor: _isProcessing
                ? Theme.of(context).colorScheme.onSurface
                : Theme.of(context).colorScheme.onPrimary,
            child: _isProcessing
                ? RotationTransition(
                    turns: _animation,
                    child: const Icon(
                      Icons.auto_awesome,
                      size: 20,
                    ),
                  )
                : widget.icon ?? const Icon(
                    Icons.auto_awesome,
                    size: 20,
                  ),
          );
        },
      ),
    );
  }
}

/// Widget to display AI enhancement status and metadata
class AIEnhancementDisplay extends StatefulWidget {
  final String noteId;

  const AIEnhancementDisplay({
    Key? key,
    required this.noteId,
  }) : super(key: key);

  @override
  State<AIEnhancementDisplay> createState() => _AIEnhancementDisplayState();
}

class _AIEnhancementDisplayState extends State<AIEnhancementDisplay> {
  Map<String, dynamic>? _enhancementData;
  bool _isLoading = true;
  String? _error;
  bool _isExpanded = true;

  @override
  void initState() {
    super.initState();
    _loadEnhancementData();
  }

  Future<void> _loadEnhancementData() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final data = await AIService().getEnhancedNote(widget.noteId);
      
      if (mounted) {
        setState(() {
          _enhancementData = data;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16.0),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
              SizedBox(width: 8),
              Text('Loading AI insights...'),
            ],
          ),
        ),
      );
    }

    if (_error != null) {
      return Card(
        color: Theme.of(context).colorScheme.errorContainer,
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Row(
            children: [
              Icon(
                Icons.error,
                color: Theme.of(context).colorScheme.onErrorContainer,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Failed to load AI insights',
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.onErrorContainer,
                  ),
                ),
              ),
              TextButton(
                onPressed: _loadEnhancementData,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    if (_enhancementData == null) {
      return const SizedBox.shrink();
    }

    // Safely access the ai_enhancement data with proper null checking
    // Handle both old and new data formats
    dynamic aiEnhancement;
    if (_enhancementData != null) {
      // Check if data has the new nested structure
      if (_enhancementData!.containsKey('ai_enhancement')) {
        aiEnhancement = _enhancementData!['ai_enhancement'];
      } else if (_enhancementData!.containsKey('summary') || _enhancementData!.containsKey('ai_tags')) {
        // Old format - data is directly in the root
        aiEnhancement = _enhancementData;
      }
    }
    
    if (aiEnhancement == null) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Row(
            children: [
              Icon(
                Icons.auto_awesome,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(width: 8),
              const Expanded(
                child: Text('AI processing pending...'),
              ),
              AIEnhancementButton(
                noteId: widget.noteId,
                onProcessingComplete: _loadEnhancementData,
              ),
            ],
          ),
        ),
      );
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.auto_awesome,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Text(
                  'AI Insights',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const Spacer(),
                IconButton(
                  icon: Icon(_isExpanded ? Icons.expand_less : Icons.expand_more),
                  onPressed: () {
                    setState(() {
                      _isExpanded = !_isExpanded;
                    });
                  },
                  tooltip: _isExpanded ? 'Collapse' : 'Expand',
                ),
                IconButton(
                  icon: const Icon(Icons.refresh),
                  onPressed: _loadEnhancementData,
                  tooltip: 'Refresh AI insights',
                ),
              ],
            ),
            
            if (_isExpanded && aiEnhancement is Map && aiEnhancement['summary'] != null) ...[
              const SizedBox(height: 12),
              Text(
                'Summary',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 4),
              Text(aiEnhancement['summary'].toString()),
            ],

            if (_isExpanded && aiEnhancement is Map && 
                aiEnhancement['ai_tags'] != null && 
                aiEnhancement['ai_tags'] is List &&
                (aiEnhancement['ai_tags'] as List).isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(
                'AI Tags',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 4),
              Wrap(
                spacing: 8.0,
                runSpacing: 4.0,
                children: (aiEnhancement['ai_tags'] as List? ?? [])
                    .where((tag) => tag != null)
                    .map<Widget>((tag) => Chip(
                          label: Text(tag.toString()),
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        ))
                    .toList(),
              ),
            ],

            if (_isExpanded && aiEnhancement is Map && 
                aiEnhancement['action_steps'] != null && 
                aiEnhancement['action_steps'] is List &&
                (aiEnhancement['action_steps'] as List).isNotEmpty) ...[ 
              const SizedBox(height: 12),
              Text(
                'Action Steps',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 4),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: (aiEnhancement['action_steps'] as List? ?? [])
                    .where((step) => step != null && step.toString().isNotEmpty)
                    .toList()
                    .asMap()
                    .entries
                    .map<Widget>((entry) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 2.0),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Container(
                                width: 20,
                                height: 20,
                                margin: const EdgeInsets.only(right: 8),
                                decoration: BoxDecoration(
                                  color: Theme.of(context).colorScheme.primary,
                                  shape: BoxShape.circle,
                                ),
                                child: Center(
                                  child: Text(
                                    '${entry.key + 1}',
                                    style: TextStyle(
                                      color: Theme.of(context).colorScheme.onPrimary,
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ),
                              Expanded(
                                child: Text(
                                  entry.value.toString(),
                                  style: Theme.of(context).textTheme.bodyMedium,
                                ),
                              ),
                            ],
                          ),
                        ))
                    .toList(),
              ),
            ],

            if (_isExpanded && aiEnhancement is Map && 
                aiEnhancement['learning_items'] != null && 
                aiEnhancement['learning_items'] is List &&
                (aiEnhancement['learning_items'] as List).isNotEmpty) ...[ 
              const SizedBox(height: 12),
              Text(
                'Learning Opportunities',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 4),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: (aiEnhancement['learning_items'] as List? ?? [])
                    .where((item) => item != null && item.toString().isNotEmpty)
                    .map<Widget>((item) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 2.0),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Icon(
                                Icons.lightbulb_outline,
                                size: 16,
                                color: Theme.of(context).colorScheme.secondary,
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  item.toString(),
                                  style: Theme.of(context).textTheme.bodyMedium,
                                ),
                              ),
                            ],
                          ),
                        ))
                    .toList(),
              ),
            ],

            if (_isExpanded && aiEnhancement is Map &&
                aiEnhancement['related_note_ids'] != null &&
                aiEnhancement['related_note_ids'] is List &&
                (aiEnhancement['related_note_ids'] as List).isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(
                'Related Notes',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const SizedBox(height: 4),
              Text(
                '${(aiEnhancement['related_note_ids'] as List? ?? []).length} related notes found',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],

            if (_isExpanded && aiEnhancement is Map && aiEnhancement['processing_status'] != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  Icon(
                    aiEnhancement['processing_status'] == 'completed'
                        ? Icons.check_circle
                        : Icons.hourglass_empty,
                    size: 16,
                    color: aiEnhancement['processing_status'] == 'completed'
                        ? Colors.green
                        : Colors.orange,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Status: ${aiEnhancement['processing_status']}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}