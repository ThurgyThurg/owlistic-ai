import 'dart:convert';
import 'package:owlistic/services/base_service.dart';
import 'package:owlistic/models/zettelkasten.dart';
import 'package:owlistic/utils/logger.dart';

/// Service for managing Zettelkasten knowledge graph operations
class ZettelkastenService extends BaseService {
  final Logger _logger = Logger('ZettelkastenService');

  /// Create a new Zettelkasten node from existing content
  Future<ZettelNode?> createNode(CreateZettelNodeInput input) async {
    try {
      final response = await authenticatedPost(
        '/zettelkasten/nodes',
        input.toJson(),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return ZettelNode.fromJson(data['node']);
      } else {
        _logger.error('Failed to create node: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error creating node', e);
      return null;
    }
  }

  /// Get all nodes with optional filtering
  Future<List<ZettelNode>> getAllNodes({ZettelSearchInput? filter}) async {
    try {
      final queryParams = <String, dynamic>{};
      
      if (filter != null) {
        if (filter.query != null) queryParams['query'] = filter.query!;
        if (filter.nodeTypes != null) {
          for (final type in filter.nodeTypes!) {
            queryParams['node_types'] = type;
          }
        }
        if (filter.tags != null) {
          for (final tag in filter.tags!) {
            queryParams['tags'] = tag;
          }
        }
        if (filter.maxDepth != null) queryParams['max_depth'] = filter.maxDepth.toString();
        if (filter.minStrength != null) queryParams['min_strength'] = filter.minStrength.toString();
      }

      final response = await authenticatedGet('/zettelkasten/nodes', queryParameters: queryParams);

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final nodesList = data['nodes'] as List<dynamic>;
        return nodesList.map((node) => ZettelNode.fromJson(node)).toList();
      } else {
        _logger.error('Failed to get nodes: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error getting nodes', e);
      return [];
    }
  }

  /// Get a specific node by ID
  Future<ZettelNode?> getNodeById(String nodeId) async {
    try {
      final response = await authenticatedGet('/zettelkasten/nodes/$nodeId');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return ZettelNode.fromJson(data['node']);
      } else if (response.statusCode == 404) {
        _logger.warning('Node not found: $nodeId');
        return null;
      } else {
        _logger.error('Failed to get node: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error getting node by ID', e);
      return null;
    }
  }

  /// Update the position of a node in the graph
  Future<bool> updateNodePosition(String nodeId, NodePosition position) async {
    try {
      final response = await authenticatedPut(
        '/zettelkasten/nodes/$nodeId/position',
        {
          'node_id': nodeId,
          'position': position.toJson(),
        },
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        _logger.error('Failed to update node position: ${response.statusCode} ${response.body}');
        return false;
      }
    } catch (e) {
      _logger.error('Error updating node position', e);
      return false;
    }
  }

  /// Delete a node from the graph
  Future<bool> deleteNode(String nodeId) async {
    try {
      final response = await authenticatedDelete('/zettelkasten/nodes/$nodeId');

      if (response.statusCode == 200) {
        return true;
      } else if (response.statusCode == 404) {
        _logger.warning('Node not found for deletion: $nodeId');
        return false;
      } else {
        _logger.error('Failed to delete node: ${response.statusCode} ${response.body}');
        return false;
      }
    } catch (e) {
      _logger.error('Error deleting node', e);
      return false;
    }
  }

  /// Create a connection between two nodes
  Future<ZettelEdge?> createConnection(CreateZettelEdgeInput input) async {
    try {
      final response = await authenticatedPost(
        '/zettelkasten/connections',
        input.toJson(),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return ZettelEdge.fromJson(data['connection']);
      } else {
        _logger.error('Failed to create connection: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error creating connection', e);
      return null;
    }
  }

  /// Delete a connection between nodes
  Future<bool> deleteConnection(String edgeId) async {
    try {
      final response = await authenticatedDelete('/zettelkasten/connections/$edgeId');

      if (response.statusCode == 200) {
        return true;
      } else if (response.statusCode == 404) {
        _logger.warning('Connection not found for deletion: $edgeId');
        return false;
      } else {
        _logger.error('Failed to delete connection: ${response.statusCode} ${response.body}');
        return false;
      }
    } catch (e) {
      _logger.error('Error deleting connection', e);
      return false;
    }
  }

  /// Get all available tags
  Future<List<ZettelTag>> getAllTags() async {
    try {
      final response = await authenticatedGet('/zettelkasten/tags');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final tagsList = data['tags'] as List<dynamic>;
        return tagsList.map((tag) => ZettelTag.fromJson(tag)).toList();
      } else {
        _logger.error('Failed to get tags: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error getting tags', e);
      return [];
    }
  }

  /// Create a new tag
  Future<ZettelTag?> createTag(CreateZettelTagInput input) async {
    try {
      final response = await authenticatedPost(
        '/zettelkasten/tags',
        input.toJson(),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return ZettelTag.fromJson(data['tag']);
      } else {
        _logger.error('Failed to create tag: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error creating tag', e);
      return null;
    }
  }

  /// Get the complete graph data for visualization
  Future<ZettelGraph?> getGraphData() async {
    try {
      final response = await authenticatedGet('/zettelkasten/graph');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return ZettelGraph.fromJson(data);
      } else {
        _logger.error('Failed to get graph data: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error getting graph data', e);
      return null;
    }
  }

  /// Export the complete graph data
  Future<ZettelGraph?> exportGraph() async {
    try {
      final response = await authenticatedGet('/zettelkasten/graph/export');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return ZettelGraph.fromJson(data);
      } else {
        _logger.error('Failed to export graph: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error exporting graph', e);
      return null;
    }
  }

  /// Request AI analysis of the knowledge graph
  Future<Map<String, dynamic>?> analyzeGraph() async {
    try {
      final response = await authenticatedPost('/zettelkasten/graph/analyze', {});

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return data['analysis'];
      } else {
        _logger.error('Failed to analyze graph: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error analyzing graph', e);
      return null;
    }
  }

  /// Search for nodes based on criteria
  Future<List<ZettelNode>> searchNodes(ZettelSearchInput input) async {
    try {
      final response = await authenticatedPost(
        '/zettelkasten/search',
        input.toJson(),
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final nodesList = data['nodes'] as List<dynamic>;
        return nodesList.map((node) => ZettelNode.fromJson(node)).toList();
      } else {
        _logger.error('Failed to search nodes: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error searching nodes', e);
      return [];
    }
  }

  /// Discover potential connections for a node
  Future<Map<String, dynamic>?> discoverConnections(String nodeId) async {
    try {
      final response = await authenticatedGet('/zettelkasten/discover/$nodeId');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return data;
      } else if (response.statusCode == 404) {
        _logger.warning('Node not found for discovery: $nodeId');
        return null;
      } else {
        _logger.error('Failed to discover connections: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error discovering connections', e);
      return null;
    }
  }

  /// Synchronize all notes to create Zettelkasten nodes
  Future<Map<String, dynamic>?> syncNotes() async {
    try {
      final response = await authenticatedPost('/zettelkasten/sync/notes', {});

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        _logger.info('Notes synchronized: ${data['message']}');
        return data;
      } else {
        _logger.error('Failed to sync notes: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error syncing notes', e);
      return null;
    }
  }

  /// Synchronize all tasks to create Zettelkasten nodes
  Future<Map<String, dynamic>?> syncTasks() async {
    try {
      final response = await authenticatedPost('/zettelkasten/sync/tasks', {});

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        _logger.info('Tasks synchronized: ${data['message']}');
        return data;
      } else {
        _logger.error('Failed to sync tasks: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error syncing tasks', e);
      return null;
    }
  }

  /// Synchronize all content (notes, tasks, projects) to create Zettelkasten nodes
  Future<Map<String, dynamic>?> syncAll() async {
    try {
      final response = await authenticatedPost('/zettelkasten/sync/all', {});

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        _logger.info('All content synchronized: ${data['message']}');
        return data;
      } else {
        _logger.error('Failed to sync all content: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error syncing all content', e);
      return null;
    }
  }

}