import 'package:flutter/material.dart';
import 'dart:math' as math;
import 'package:owlistic/models/zettelkasten.dart';
import 'package:owlistic/utils/logger.dart';

/// Interactive graph visualization widget for the Zettelkasten knowledge graph
class ZettelkastenGraphWidget extends StatefulWidget {
  final ZettelGraph graph;
  final ZettelNode? selectedNode;
  final Function(ZettelNode?) onNodeSelected;
  final Function(String nodeId, NodePosition position) onNodePositionChanged;
  final Function(String sourceId, String targetId, String connectionType) onConnectionCreated;
  final Function(String edgeId) onConnectionDeleted;

  const ZettelkastenGraphWidget({
    super.key,
    required this.graph,
    this.selectedNode,
    required this.onNodeSelected,
    required this.onNodePositionChanged,
    required this.onConnectionCreated,
    required this.onConnectionDeleted,
  });

  @override
  State<ZettelkastenGraphWidget> createState() => _ZettelkastenGraphWidgetState();
}

class _ZettelkastenGraphWidgetState extends State<ZettelkastenGraphWidget>
    with TickerProviderStateMixin {
  final Logger _logger = Logger('ZettelkastenGraphWidget');

  // Transform state
  late TransformationController _transformationController;
  
  // Node positioning
  final Map<String, Offset> _nodePositions = {};
  final Map<String, Size> _nodeSizes = {};
  
  // Interaction state
  String? _draggedNodeId;
  Offset? _dragOffset;
  bool _isCreatingConnection = false;
  String? _connectionSourceId;
  Offset? _connectionEndPoint;

  // Animation
  late AnimationController _layoutAnimationController;
  late Animation<double> _layoutAnimation;

  // Constants
  static const double _nodeRadius = 25.0;
  static const double _minZoom = 0.1;
  static const double _maxZoom = 3.0;
  static const double _connectionLineWidth = 2.0;
  static const double _selectedNodeBorder = 3.0;

  @override
  void initState() {
    super.initState();
    _transformationController = TransformationController();
    
    _layoutAnimationController = AnimationController(
      duration: const Duration(milliseconds: 1000),
      vsync: this,
    );
    
    _layoutAnimation = CurvedAnimation(
      parent: _layoutAnimationController,
      curve: Curves.easeInOut,
    );

    _initializeNodePositions();
    _startLayoutAnimation();
  }

  @override
  void didUpdateWidget(ZettelkastenGraphWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    
    if (widget.graph.nodes.length != oldWidget.graph.nodes.length ||
        widget.graph.edges.length != oldWidget.graph.edges.length) {
      _updateNodePositions();
      _startLayoutAnimation();
    }
  }

  @override
  void dispose() {
    _transformationController.dispose();
    _layoutAnimationController.dispose();
    super.dispose();
  }

  void _initializeNodePositions() {
    if (widget.graph.nodes.isEmpty) return;

    final center = const Offset(400, 300);
    final radius = 200.0;
    
    for (int i = 0; i < widget.graph.nodes.length; i++) {
      final node = widget.graph.nodes[i];
      
      // Use saved position if available, otherwise create a new one
      if (node.position != null) {
        _nodePositions[node.id] = Offset(node.position!.x, node.position!.y);
      } else {
        // Arrange nodes in a circle initially
        final angle = (2 * math.pi * i) / widget.graph.nodes.length;
        final x = center.dx + radius * math.cos(angle);
        final y = center.dy + radius * math.sin(angle);
        _nodePositions[node.id] = Offset(x, y);
      }
      
      // Set node size
      _nodeSizes[node.id] = const Size(_nodeRadius * 2, _nodeRadius * 2);
    }
  }

  void _updateNodePositions() {
    // Add positions for new nodes
    final existingNodeIds = _nodePositions.keys.toSet();
    final newNodes = widget.graph.nodes.where((node) => !existingNodeIds.contains(node.id));
    
    if (newNodes.isNotEmpty) {
      final center = const Offset(400, 300);
      final radius = 250.0;
      
      for (final node in newNodes) {
        if (node.position != null) {
          _nodePositions[node.id] = Offset(node.position!.x, node.position!.y);
        } else {
          // Place new nodes randomly around the center
          final angle = math.Random().nextDouble() * 2 * math.pi;
          final distance = radius + math.Random().nextDouble() * 100;
          final x = center.dx + distance * math.cos(angle);
          final y = center.dy + distance * math.sin(angle);
          _nodePositions[node.id] = Offset(x, y);
        }
        _nodeSizes[node.id] = const Size(_nodeRadius * 2, _nodeRadius * 2);
      }
    }
    
    // Remove positions for deleted nodes
    final currentNodeIds = widget.graph.nodes.map((node) => node.id).toSet();
    _nodePositions.removeWhere((id, _) => !currentNodeIds.contains(id));
    _nodeSizes.removeWhere((id, _) => !currentNodeIds.contains(id));
  }

  void _startLayoutAnimation() {
    _layoutAnimationController.reset();
    _layoutAnimationController.forward();
  }

  Color _getNodeColor(ZettelNode node) {
    switch (node.nodeType) {
      case 'note':
        return Colors.blue.shade400;
      case 'task':
        return Colors.green.shade400;
      case 'project':
        return Colors.orange.shade400;
      default:
        return Colors.grey.shade400;
    }
  }

  Color _getConnectionColor(ZettelEdge edge) {
    switch (edge.connectionType) {
      case 'related':
        return Colors.grey.shade600;
      case 'depends_on':
        return Colors.red.shade600;
      case 'references':
        return Colors.blue.shade600;
      case 'supports':
        return Colors.green.shade600;
      case 'contradicts':
        return Colors.orange.shade600;
      default:
        return Colors.grey.shade600;
    }
  }

  void _handleNodeTap(ZettelNode node) {
    widget.onNodeSelected(widget.selectedNode?.id == node.id ? null : node);
  }

  void _handleNodePanStart(ZettelNode node, DragStartDetails details) {
    _draggedNodeId = node.id;
    final nodePosition = _nodePositions[node.id] ?? Offset.zero;
    _dragOffset = details.localPosition - nodePosition;
  }

  void _handleNodePanUpdate(DragUpdateDetails details) {
    if (_draggedNodeId != null && _dragOffset != null) {
      setState(() {
        _nodePositions[_draggedNodeId!] = details.localPosition - _dragOffset!;
      });
    }
  }

  void _handleNodePanEnd(DragEndDetails details) {
    if (_draggedNodeId != null) {
      final position = _nodePositions[_draggedNodeId!] ?? Offset.zero;
      widget.onNodePositionChanged(
        _draggedNodeId!,
        NodePosition(x: position.dx, y: position.dy),
      );
      _draggedNodeId = null;
      _dragOffset = null;
    }
  }

  void _handleCanvasTap(TapUpDetails details) {
    // Deselect node if clicking on empty space
    widget.onNodeSelected(null);
  }

  Offset _getNodePosition(String nodeId) {
    return _nodePositions[nodeId] ?? Offset.zero;
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: double.infinity,
      color: Theme.of(context).colorScheme.surface,
      child: InteractiveViewer(
        transformationController: _transformationController,
        minScale: _minZoom,
        maxScale: _maxZoom,
        boundaryMargin: const EdgeInsets.all(100),
        child: GestureDetector(
          onTapUp: _handleCanvasTap,
          child: CustomPaint(
            size: const Size(1200, 800),
            painter: _GraphPainter(
              graph: widget.graph,
              nodePositions: _nodePositions,
              selectedNodeId: widget.selectedNode?.id,
              getNodeColor: _getNodeColor,
              getConnectionColor: _getConnectionColor,
              animation: _layoutAnimation,
            ),
            child: Stack(
              children: [
                // Nodes
                ...widget.graph.nodes.map((node) {
                  final position = _getNodePosition(node.id);
                  final isSelected = widget.selectedNode?.id == node.id;
                  
                  return Positioned(
                    left: position.dx - _nodeRadius,
                    top: position.dy - _nodeRadius,
                    child: GestureDetector(
                      onTap: () => _handleNodeTap(node),
                      onPanStart: (details) => _handleNodePanStart(node, details),
                      onPanUpdate: _handleNodePanUpdate,
                      onPanEnd: _handleNodePanEnd,
                      child: AnimatedBuilder(
                        animation: _layoutAnimation,
                        builder: (context, child) {
                          return Transform.scale(
                            scale: _layoutAnimation.value,
                            child: Container(
                              width: _nodeRadius * 2,
                              height: _nodeRadius * 2,
                              decoration: BoxDecoration(
                                color: _getNodeColor(node),
                                shape: BoxShape.circle,
                                border: isSelected
                                    ? Border.all(
                                        color: Theme.of(context).colorScheme.primary,
                                        width: _selectedNodeBorder,
                                      )
                                    : null,
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withOpacity(0.2),
                                    blurRadius: 4,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: Center(
                                child: Text(
                                  _getNodeIcon(node.nodeType),
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                  );
                }),
                
                // Node labels
                ...widget.graph.nodes.map((node) {
                  final position = _getNodePosition(node.id);
                  
                  return Positioned(
                    left: position.dx - 60,
                    top: position.dy + _nodeRadius + 5,
                    child: AnimatedBuilder(
                      animation: _layoutAnimation,
                      builder: (context, child) {
                        return Opacity(
                          opacity: _layoutAnimation.value,
                          child: Container(
                            width: 120,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 4,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.black.withOpacity(0.7),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              node.title,
                              textAlign: TextAlign.center,
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 10,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  );
                }),
              ],
            ),
          ),
        ),
      ),
    );
  }

  String _getNodeIcon(String nodeType) {
    switch (nodeType) {
      case 'note':
        return '📝';
      case 'task':
        return '✅';
      case 'project':
        return '📁';
      default:
        return '●';
    }
  }
}

/// Custom painter for drawing the graph connections
class _GraphPainter extends CustomPainter {
  final ZettelGraph graph;
  final Map<String, Offset> nodePositions;
  final String? selectedNodeId;
  final Color Function(ZettelNode) getNodeColor;
  final Color Function(ZettelEdge) getConnectionColor;
  final Animation<double> animation;

  _GraphPainter({
    required this.graph,
    required this.nodePositions,
    this.selectedNodeId,
    required this.getNodeColor,
    required this.getConnectionColor,
    required this.animation,
  }) : super(repaint: animation);

  @override
  void paint(Canvas canvas, Size size) {
    // Draw connections
    for (final edge in graph.edges) {
      final sourcePos = nodePositions[edge.sourceNodeId];
      final targetPos = nodePositions[edge.targetNodeId];
      
      if (sourcePos != null && targetPos != null) {
        _drawConnection(canvas, edge, sourcePos, targetPos);
      }
    }
  }

  void _drawConnection(Canvas canvas, ZettelEdge edge, Offset start, Offset end) {
    final paint = Paint()
      ..color = getConnectionColor(edge).withOpacity(0.7 * animation.value)
      ..strokeWidth = 2.0 * edge.strength
      ..style = PaintingStyle.stroke;

    // Draw different line styles based on connection type
    switch (edge.connectionType) {
      case 'depends_on':
        _drawArrowLine(canvas, paint, start, end);
        break;
      case 'contradicts':
        _drawDashedLine(canvas, paint, start, end);
        break;
      default:
        canvas.drawLine(start, end, paint);
        break;
    }

    // Draw connection strength indicator
    if (edge.strength > 0.7) {
      final midPoint = Offset(
        (start.dx + end.dx) / 2,
        (start.dy + end.dy) / 2,
      );
      
      final strengthPaint = Paint()
        ..color = getConnectionColor(edge)
        ..style = PaintingStyle.fill;
      
      canvas.drawCircle(midPoint, 3.0 * animation.value, strengthPaint);
    }
  }

  void _drawArrowLine(Canvas canvas, Paint paint, Offset start, Offset end) {
    canvas.drawLine(start, end, paint);
    
    // Draw arrowhead
    final direction = (end - start).normalize();
    final arrowHead1 = end - (direction * 10) + (direction.rotate(0.5) * 5);
    final arrowHead2 = end - (direction * 10) - (direction.rotate(0.5) * 5);
    
    final arrowPath = Path()
      ..moveTo(end.dx, end.dy)
      ..lineTo(arrowHead1.dx, arrowHead1.dy)
      ..lineTo(arrowHead2.dx, arrowHead2.dy)
      ..close();
    
    paint.style = PaintingStyle.fill;
    canvas.drawPath(arrowPath, paint);
    paint.style = PaintingStyle.stroke;
  }

  void _drawDashedLine(Canvas canvas, Paint paint, Offset start, Offset end) {
    const dashLength = 5.0;
    const gapLength = 3.0;
    
    final direction = (end - start).normalize();
    final distance = (end - start).distance;
    
    double currentDistance = 0;
    Offset currentPoint = start;
    
    while (currentDistance < distance) {
      final dashEnd = currentDistance + dashLength;
      final nextPoint = start + (direction * math.min(dashEnd, distance));
      
      canvas.drawLine(currentPoint, nextPoint, paint);
      
      currentDistance = dashEnd + gapLength;
      currentPoint = start + (direction * currentDistance);
    }
  }

  @override
  bool shouldRepaint(covariant _GraphPainter oldDelegate) {
    return oldDelegate.graph != graph ||
           oldDelegate.nodePositions != nodePositions ||
           oldDelegate.selectedNodeId != selectedNodeId;
  }
}

/// Extension methods for vector operations
extension OffsetExtension on Offset {
  Offset normalize() {
    final length = distance;
    if (length == 0) return Offset.zero;
    return this / length;
  }
  
  Offset rotate(double angle) {
    final cos = math.cos(angle);
    final sin = math.sin(angle);
    return Offset(
      dx * cos - dy * sin,
      dx * sin + dy * cos,
    );
  }
}