import 'package:flutter/foundation.dart';
import 'package:owlistic/models/zettelkasten.dart';
import 'package:owlistic/services/zettelkasten_service.dart';
import 'package:owlistic/utils/logger.dart';

/// Provider for managing Zettelkasten knowledge graph state
class ZettelkastenProvider with ChangeNotifier {
  final Logger _logger = Logger('ZettelkastenProvider');
  final ZettelkastenService _service;

  ZettelkastenProvider({required ZettelkastenService service})
      : _service = service;

  // State
  ZettelGraph? _graph;
  ZettelGraph? _filteredGraph;
  bool _isLoading = false;
  String? _errorMessage;
  ZettelSearchInput? _currentFilter;

  // Getters
  ZettelGraph? get graph => _graph;
  ZettelGraph? get filteredGraph => _filteredGraph;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;
  ZettelSearchInput? get currentFilter => _currentFilter;

  /// Load the complete graph data
  Future<void> loadGraphData() async {
    _setLoading(true);
    _clearError();

    try {
      _logger.info('Loading graph data...');
      final graph = await _service.getGraphData();
      
      if (graph != null) {
        _graph = graph;
        _applyCurrentFilter();
        _logger.info('Graph loaded: ${graph.nodes.length} nodes, ${graph.edges.length} edges');
      } else {
        _setError('Failed to load graph data');
      }
    } catch (e) {
      _logger.error('Error loading graph data', e);
      _setError('Error loading graph: $e');
    } finally {
      _setLoading(false);
    }
  }

  /// Create a new node from existing content
  Future<ZettelNode?> createNode(CreateZettelNodeInput input) async {
    try {
      _logger.info('Creating node: ${input.title}');
      final node = await _service.createNode(input);
      
      if (node != null) {
        // Add to current graph
        if (_graph != null) {
          _graph = _graph!.copyWith(
            nodes: [..._graph!.nodes, node],
          );
          _applyCurrentFilter();
        }
        _logger.info('Node created successfully: ${node.id}');
      }
      
      return node;
    } catch (e) {
      _logger.error('Error creating node', e);
      return null;
    }
  }

  /// Update the position of a node
  Future<bool> updateNodePosition(String nodeId, NodePosition position) async {
    try {
      final success = await _service.updateNodePosition(nodeId, position);
      
      if (success && _graph != null) {
        // Update the node position in the local graph
        final updatedNodes = _graph!.nodes.map((node) {
          if (node.id == nodeId) {
            return node.copyWith(position: position);
          }
          return node;
        }).toList();
        
        _graph = _graph!.copyWith(nodes: updatedNodes);
        _applyCurrentFilter();
      }
      
      return success;
    } catch (e) {
      _logger.error('Error updating node position', e);
      return false;
    }
  }

  /// Create a connection between two nodes
  Future<ZettelEdge?> createConnection(
    String sourceNodeId, 
    String targetNodeId, 
    String connectionType,
    {String? description, double? strength}
  ) async {
    try {
      final input = CreateZettelEdgeInput(
        sourceNodeId: sourceNodeId,
        targetNodeId: targetNodeId,
        connectionType: connectionType,
        description: description,
        strength: strength,
        isAutomatic: false, // User-created connection
      );
      
      _logger.info('Creating connection: $sourceNodeId -> $targetNodeId ($connectionType)');
      final edge = await _service.createConnection(input);
      
      if (edge != null && _graph != null) {
        // Add to current graph
        _graph = _graph!.copyWith(
          edges: [..._graph!.edges, edge],
        );
        _applyCurrentFilter();
        _logger.info('Connection created successfully: ${edge.id}');
      }
      
      return edge;
    } catch (e) {
      _logger.error('Error creating connection', e);
      return null;
    }
  }

  /// Delete a connection
  Future<bool> deleteConnection(String edgeId) async {
    try {
      _logger.info('Deleting connection: $edgeId');
      final success = await _service.deleteConnection(edgeId);
      
      if (success && _graph != null) {
        // Remove from current graph
        final updatedEdges = _graph!.edges.where((edge) => edge.id != edgeId).toList();
        _graph = _graph!.copyWith(edges: updatedEdges);
        _applyCurrentFilter();
        _logger.info('Connection deleted successfully');
      }
      
      return success;
    } catch (e) {
      _logger.error('Error deleting connection', e);
      return false;
    }
  }

  /// Delete a node and its connections
  Future<bool> deleteNode(String nodeId) async {
    try {
      _logger.info('Deleting node: $nodeId');
      final success = await _service.deleteNode(nodeId);
      
      if (success && _graph != null) {
        // Remove node and its connections from current graph
        final updatedNodes = _graph!.nodes.where((node) => node.id != nodeId).toList();
        final updatedEdges = _graph!.edges.where((edge) => 
          edge.sourceNodeId != nodeId && edge.targetNodeId != nodeId
        ).toList();
        
        _graph = _graph!.copyWith(
          nodes: updatedNodes,
          edges: updatedEdges,
        );
        _applyCurrentFilter();
        _logger.info('Node deleted successfully');
      }
      
      return success;
    } catch (e) {
      _logger.error('Error deleting node', e);
      return false;
    }
  }

  /// Create a new tag
  Future<ZettelTag?> createTag(CreateZettelTagInput input) async {
    try {
      _logger.info('Creating tag: ${input.name}');
      final tag = await _service.createTag(input);
      
      if (tag != null && _graph != null) {
        // Add to current graph
        _graph = _graph!.copyWith(
          tags: [..._graph!.tags, tag],
        );
        notifyListeners();
        _logger.info('Tag created successfully: ${tag.id}');
      }
      
      return tag;
    } catch (e) {
      _logger.error('Error creating tag', e);
      return null;
    }
  }

  /// Apply a filter to the graph
  void applyFilter(ZettelSearchInput? filter) {
    _currentFilter = filter;
    _applyCurrentFilter();
  }

  /// Clear current filters
  void clearFilter() {
    _currentFilter = null;
    _filteredGraph = null;
    notifyListeners();
  }

  /// Search for nodes
  Future<List<ZettelNode>> searchNodes(ZettelSearchInput input) async {
    try {
      _logger.info('Searching nodes: ${input.query}');
      return await _service.searchNodes(input);
    } catch (e) {
      _logger.error('Error searching nodes', e);
      return [];
    }
  }

  /// Discover potential connections for a node
  Future<Map<String, dynamic>?> discoverConnections(String nodeId) async {
    try {
      _logger.info('Discovering connections for node: $nodeId');
      return await _service.discoverConnections(nodeId);
    } catch (e) {
      _logger.error('Error discovering connections', e);
      return null;
    }
  }

  /// Synchronize all content to create nodes
  Future<Map<String, dynamic>?> syncAllContent() async {
    try {
      _logger.info('Synchronizing all content...');
      return await _service.syncAll();
    } catch (e) {
      _logger.error('Error syncing all content', e);
      return null;
    }
  }

  /// Synchronize notes only
  Future<Map<String, dynamic>?> syncNotes() async {
    try {
      _logger.info('Synchronizing notes...');
      return await _service.syncNotes();
    } catch (e) {
      _logger.error('Error syncing notes', e);
      return null;
    }
  }

  /// Synchronize tasks only
  Future<Map<String, dynamic>?> syncTasks() async {
    try {
      _logger.info('Synchronizing tasks...');
      return await _service.syncTasks();
    } catch (e) {
      _logger.error('Error syncing tasks', e);
      return null;
    }
  }

  /// Export the complete graph
  Future<ZettelGraph?> exportGraph() async {
    try {
      _logger.info('Exporting graph...');
      return await _service.exportGraph();
    } catch (e) {
      _logger.error('Error exporting graph', e);
      return null;
    }
  }

  /// Analyze the knowledge graph using AI
  Future<Map<String, dynamic>?> analyzeGraph() async {
    try {
      _logger.info('Analyzing graph...');
      return await _service.analyzeGraph();
    } catch (e) {
      _logger.error('Error analyzing graph', e);
      return null;
    }
  }

  /// Get a specific node by ID
  Future<ZettelNode?> getNodeById(String nodeId) async {
    try {
      return await _service.getNodeById(nodeId);
    } catch (e) {
      _logger.error('Error getting node by ID', e);
      return null;
    }
  }

  /// Get all available tags
  Future<List<ZettelTag>> getAllTags() async {
    try {
      return await _service.getAllTags();
    } catch (e) {
      _logger.error('Error getting all tags', e);
      return [];
    }
  }

  // Private methods

  void _applyCurrentFilter() {
    if (_graph == null) {
      _filteredGraph = null;
      notifyListeners();
      return;
    }

    if (_currentFilter == null) {
      _filteredGraph = null;
      notifyListeners();
      return;
    }

    List<ZettelNode> filteredNodes = _graph!.nodes;

    // Apply node type filter
    if (_currentFilter!.nodeTypes != null && _currentFilter!.nodeTypes!.isNotEmpty) {
      filteredNodes = filteredNodes.where((node) => 
        _currentFilter!.nodeTypes!.contains(node.nodeType)
      ).toList();
    }

    // Apply tag filter
    if (_currentFilter!.tags != null && _currentFilter!.tags!.isNotEmpty) {
      filteredNodes = filteredNodes.where((node) {
        final nodeTagNames = node.tagNames;
        return _currentFilter!.tags!.any((filterTag) => 
          nodeTagNames.contains(filterTag)
        );
      }).toList();
    }

    // Apply text search filter
    if (_currentFilter!.query != null && _currentFilter!.query!.isNotEmpty) {
      final query = _currentFilter!.query!.toLowerCase();
      filteredNodes = filteredNodes.where((node) =>
        node.title.toLowerCase().contains(query) ||
        (node.summary?.toLowerCase().contains(query) ?? false) ||
        node.tagNames.any((tag) => tag.toLowerCase().contains(query))
      ).toList();
    }

    // Filter edges to only include connections between filtered nodes
    final filteredNodeIds = filteredNodes.map((node) => node.id).toSet();
    final filteredEdges = _graph!.edges.where((edge) =>
      filteredNodeIds.contains(edge.sourceNodeId) &&
      filteredNodeIds.contains(edge.targetNodeId)
    ).toList();

    // Apply strength filter
    List<ZettelEdge> strengthFilteredEdges = filteredEdges;
    if (_currentFilter!.minStrength != null) {
      strengthFilteredEdges = filteredEdges.where((edge) =>
        edge.strength >= _currentFilter!.minStrength!
      ).toList();
    }

    _filteredGraph = ZettelGraph(
      nodes: filteredNodes,
      edges: strengthFilteredEdges,
      tags: _graph!.tags,
    );

    notifyListeners();
  }

  void _setLoading(bool loading) {
    _isLoading = loading;
    notifyListeners();
  }

  void _setError(String error) {
    _errorMessage = error;
    notifyListeners();
  }

  void _clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  /// Get statistics about the current graph
  Map<String, dynamic> getGraphStats() {
    if (_graph == null) {
      return {
        'totalNodes': 0,
        'totalEdges': 0,
        'totalTags': 0,
        'nodesByType': <String, int>{},
        'edgesByType': <String, int>{},
        'tagsByCategory': <String, int>{},
      };
    }

    final nodesByType = <String, int>{};
    for (final node in _graph!.nodes) {
      nodesByType[node.nodeType] = (nodesByType[node.nodeType] ?? 0) + 1;
    }

    final edgesByType = <String, int>{};
    for (final edge in _graph!.edges) {
      edgesByType[edge.connectionType] = (edgesByType[edge.connectionType] ?? 0) + 1;
    }

    final tagsByCategory = <String, int>{};
    for (final tag in _graph!.tags) {
      final category = tag.category ?? 'uncategorized';
      tagsByCategory[category] = (tagsByCategory[category] ?? 0) + 1;
    }

    return {
      'totalNodes': _graph!.nodes.length,
      'totalEdges': _graph!.edges.length,
      'totalTags': _graph!.tags.length,
      'nodesByType': nodesByType,
      'edgesByType': edgesByType,
      'tagsByCategory': tagsByCategory,
    };
  }
}