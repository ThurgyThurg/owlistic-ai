import 'package:uuid/uuid.dart';

/// Represents a position in 2D space for graph visualization
class NodePosition {
  final double x;
  final double y;

  const NodePosition({required this.x, required this.y});

  factory NodePosition.fromJson(Map<String, dynamic> json) {
    return NodePosition(
      x: (json['x'] as num?)?.toDouble() ?? 0.0,
      y: (json['y'] as num?)?.toDouble() ?? 0.0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'x': x,
      'y': y,
    };
  }

  NodePosition copyWith({double? x, double? y}) {
    return NodePosition(
      x: x ?? this.x,
      y: y ?? this.y,
    );
  }

  @override
  String toString() => 'NodePosition(x: $x, y: $y)';

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is NodePosition && other.x == x && other.y == y;
  }

  @override
  int get hashCode => x.hashCode ^ y.hashCode;
}

/// Represents a tag in the Zettelkasten system
class ZettelTag {
  final String id;
  final String name;
  final String? description;
  final String color;
  final String? category;
  final DateTime createdAt;
  final DateTime updatedAt;

  const ZettelTag({
    required this.id,
    required this.name,
    this.description,
    required this.color,
    this.category,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ZettelTag.fromJson(Map<String, dynamic> json) {
    return ZettelTag(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      color: json['color'] as String? ?? '#3b82f6',
      category: json['category'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'description': description,
      'color': color,
      'category': category,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  ZettelTag copyWith({
    String? id,
    String? name,
    String? description,
    String? color,
    String? category,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ZettelTag(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      color: color ?? this.color,
      category: category ?? this.category,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  String toString() => 'ZettelTag(id: $id, name: $name, category: $category)';
}

/// Represents a node in the Zettelkasten graph
class ZettelNode {
  final String id;
  final String nodeType; // "note", "task", "project"
  final String nodeId; // ID of the referenced content
  final String title;
  final String? summary;
  final List<ZettelTag> tags;
  final List<ZettelEdge> connections;
  final List<ZettelEdge> backLinks;
  final NodePosition? position;
  final DateTime createdAt;
  final DateTime updatedAt;

  const ZettelNode({
    required this.id,
    required this.nodeType,
    required this.nodeId,
    required this.title,
    this.summary,
    required this.tags,
    required this.connections,
    required this.backLinks,
    this.position,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ZettelNode.fromJson(Map<String, dynamic> json) {
    return ZettelNode(
      id: json['id'] as String,
      nodeType: json['node_type'] as String,
      nodeId: json['node_id'] as String,
      title: json['title'] as String,
      summary: json['summary'] as String?,
      tags: (json['tags'] as List<dynamic>?)
          ?.map((tag) => ZettelTag.fromJson(tag as Map<String, dynamic>))
          .toList() ?? [],
      connections: (json['outgoing_connections'] as List<dynamic>?)
          ?.map((conn) => ZettelEdge.fromJson(conn as Map<String, dynamic>))
          .toList() ?? [],
      backLinks: (json['incoming_connections'] as List<dynamic>?)
          ?.map((conn) => ZettelEdge.fromJson(conn as Map<String, dynamic>))
          .toList() ?? [],
      position: json['position'] != null 
          ? NodePosition.fromJson(json['position'] as Map<String, dynamic>)
          : null,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'node_type': nodeType,
      'node_id': nodeId,
      'title': title,
      'summary': summary,
      'tags': tags.map((tag) => tag.toJson()).toList(),
      'outgoing_connections': connections.map((conn) => conn.toJson()).toList(),
      'incoming_connections': backLinks.map((conn) => conn.toJson()).toList(),
      'position': position?.toJson(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Get all connected node IDs (both outgoing and incoming)
  List<String> get connectedNodeIds {
    final Set<String> connected = {};
    for (final connection in connections) {
      connected.add(connection.targetNodeId);
    }
    for (final backLink in backLinks) {
      connected.add(backLink.sourceNodeId);
    }
    return connected.toList();
  }

  /// Get tag names as a list of strings
  List<String> get tagNames => tags.map((tag) => tag.name).toList();

  ZettelNode copyWith({
    String? id,
    String? nodeType,
    String? nodeId,
    String? title,
    String? summary,
    List<ZettelTag>? tags,
    List<ZettelEdge>? connections,
    List<ZettelEdge>? backLinks,
    NodePosition? position,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ZettelNode(
      id: id ?? this.id,
      nodeType: nodeType ?? this.nodeType,
      nodeId: nodeId ?? this.nodeId,
      title: title ?? this.title,
      summary: summary ?? this.summary,
      tags: tags ?? this.tags,
      connections: connections ?? this.connections,
      backLinks: backLinks ?? this.backLinks,
      position: position ?? this.position,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  String toString() => 'ZettelNode(id: $id, type: $nodeType, title: $title)';
}

/// Represents a connection between two nodes
class ZettelEdge {
  final String id;
  final String sourceNodeId;
  final String targetNodeId;
  final ZettelNode? sourceNode;
  final ZettelNode? targetNode;
  final String connectionType; // "related", "depends_on", "references", "contradicts", "supports"
  final double strength;
  final String? description;
  final bool isAutomatic;
  final DateTime createdAt;
  final DateTime updatedAt;

  const ZettelEdge({
    required this.id,
    required this.sourceNodeId,
    required this.targetNodeId,
    this.sourceNode,
    this.targetNode,
    required this.connectionType,
    required this.strength,
    this.description,
    required this.isAutomatic,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ZettelEdge.fromJson(Map<String, dynamic> json) {
    return ZettelEdge(
      id: json['id'] as String,
      sourceNodeId: json['source_node_id'] as String,
      targetNodeId: json['target_node_id'] as String,
      sourceNode: json['source_node'] != null
          ? ZettelNode.fromJson(json['source_node'] as Map<String, dynamic>)
          : null,
      targetNode: json['target_node'] != null
          ? ZettelNode.fromJson(json['target_node'] as Map<String, dynamic>)
          : null,
      connectionType: json['connection_type'] as String,
      strength: (json['strength'] as num?)?.toDouble() ?? 1.0,
      description: json['description'] as String?,
      isAutomatic: json['is_automatic'] as bool? ?? true,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'source_node_id': sourceNodeId,
      'target_node_id': targetNodeId,
      'source_node': sourceNode?.toJson(),
      'target_node': targetNode?.toJson(),
      'connection_type': connectionType,
      'strength': strength,
      'description': description,
      'is_automatic': isAutomatic,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  ZettelEdge copyWith({
    String? id,
    String? sourceNodeId,
    String? targetNodeId,
    ZettelNode? sourceNode,
    ZettelNode? targetNode,
    String? connectionType,
    double? strength,
    String? description,
    bool? isAutomatic,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ZettelEdge(
      id: id ?? this.id,
      sourceNodeId: sourceNodeId ?? this.sourceNodeId,
      targetNodeId: targetNodeId ?? this.targetNodeId,
      sourceNode: sourceNode ?? this.sourceNode,
      targetNode: targetNode ?? this.targetNode,
      connectionType: connectionType ?? this.connectionType,
      strength: strength ?? this.strength,
      description: description ?? this.description,
      isAutomatic: isAutomatic ?? this.isAutomatic,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  String toString() => 'ZettelEdge(id: $id, type: $connectionType, $sourceNodeId -> $targetNodeId)';
}

/// Represents a view state for the graph visualization
class GraphViewState {
  final double centerX;
  final double centerY;
  final double zoom;

  const GraphViewState({
    required this.centerX,
    required this.centerY,
    required this.zoom,
  });

  factory GraphViewState.fromJson(Map<String, dynamic> json) {
    return GraphViewState(
      centerX: (json['center_x'] as num?)?.toDouble() ?? 0.0,
      centerY: (json['center_y'] as num?)?.toDouble() ?? 0.0,
      zoom: (json['zoom'] as num?)?.toDouble() ?? 1.0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'center_x': centerX,
      'center_y': centerY,
      'zoom': zoom,
    };
  }

  GraphViewState copyWith({
    double? centerX,
    double? centerY,
    double? zoom,
  }) {
    return GraphViewState(
      centerX: centerX ?? this.centerX,
      centerY: centerY ?? this.centerY,
      zoom: zoom ?? this.zoom,
    );
  }

  @override
  String toString() => 'GraphViewState(center: ($centerX, $centerY), zoom: $zoom)';
}

/// Represents the complete graph data
class ZettelGraph {
  final List<ZettelNode> nodes;
  final List<ZettelEdge> edges;
  final List<ZettelTag> tags;

  const ZettelGraph({
    required this.nodes,
    required this.edges,
    required this.tags,
  });

  factory ZettelGraph.fromJson(Map<String, dynamic> json) {
    return ZettelGraph(
      nodes: (json['nodes'] as List<dynamic>?)
          ?.map((node) => ZettelNode.fromJson(node as Map<String, dynamic>))
          .toList() ?? [],
      edges: (json['edges'] as List<dynamic>?)
          ?.map((edge) => ZettelEdge.fromJson(edge as Map<String, dynamic>))
          .toList() ?? [],
      tags: (json['tags'] as List<dynamic>?)
          ?.map((tag) => ZettelTag.fromJson(tag as Map<String, dynamic>))
          .toList() ?? [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'nodes': nodes.map((node) => node.toJson()).toList(),
      'edges': edges.map((edge) => edge.toJson()).toList(),
      'tags': tags.map((tag) => tag.toJson()).toList(),
    };
  }

  ZettelGraph copyWith({
    List<ZettelNode>? nodes,
    List<ZettelEdge>? edges,
    List<ZettelTag>? tags,
  }) {
    return ZettelGraph(
      nodes: nodes ?? this.nodes,
      edges: edges ?? this.edges,
      tags: tags ?? this.tags,
    );
  }

  @override
  String toString() => 'ZettelGraph(nodes: ${nodes.length}, edges: ${edges.length}, tags: ${tags.length})';
}

/// Input classes for API calls

class CreateZettelNodeInput {
  final String nodeType;
  final String nodeId;
  final String title;
  final String? summary;
  final List<String>? tags;
  final NodePosition? position;

  const CreateZettelNodeInput({
    required this.nodeType,
    required this.nodeId,
    required this.title,
    this.summary,
    this.tags,
    this.position,
  });

  Map<String, dynamic> toJson() {
    return {
      'node_type': nodeType,
      'node_id': nodeId,
      'title': title,
      'summary': summary,
      'tags': tags,
      'position': position?.toJson(),
    };
  }
}

class CreateZettelEdgeInput {
  final String sourceNodeId;
  final String targetNodeId;
  final String connectionType;
  final double? strength;
  final String? description;
  final bool? isAutomatic;

  const CreateZettelEdgeInput({
    required this.sourceNodeId,
    required this.targetNodeId,
    required this.connectionType,
    this.strength,
    this.description,
    this.isAutomatic,
  });

  Map<String, dynamic> toJson() {
    return {
      'source_node_id': sourceNodeId,
      'target_node_id': targetNodeId,
      'connection_type': connectionType,
      'strength': strength,
      'description': description,
      'is_automatic': isAutomatic,
    };
  }
}

class CreateZettelTagInput {
  final String name;
  final String? description;
  final String? color;
  final String? category;

  const CreateZettelTagInput({
    required this.name,
    this.description,
    this.color,
    this.category,
  });

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'description': description,
      'color': color,
      'category': category,
    };
  }
}

class ZettelSearchInput {
  final String? query;
  final List<String>? tags;
  final List<String>? nodeTypes;
  final int? maxDepth;
  final double? minStrength;

  const ZettelSearchInput({
    this.query,
    this.tags,
    this.nodeTypes,
    this.maxDepth,
    this.minStrength,
  });

  Map<String, dynamic> toJson() {
    return {
      'query': query,
      'tags': tags,
      'node_types': nodeTypes,
      'max_depth': maxDepth,
      'min_strength': minStrength,
    };
  }
}

/// Connection types for Zettelkasten edges
class ConnectionType {
  static const String related = 'related';
  static const String dependsOn = 'depends_on';
  static const String references = 'references';
  static const String contradicts = 'contradicts';
  static const String supports = 'supports';

  static const List<String> all = [
    related,
    dependsOn,
    references,
    contradicts,
    supports,
  ];

  static String getDisplayName(String type) {
    switch (type) {
      case related:
        return 'Related';
      case dependsOn:
        return 'Depends On';
      case references:
        return 'References';
      case contradicts:
        return 'Contradicts';
      case supports:
        return 'Supports';
      default:
        return type;
    }
  }
}

/// Node types for Zettelkasten nodes
class NodeType {
  static const String note = 'note';
  static const String task = 'task';
  static const String project = 'project';

  static const List<String> all = [note, task, project];

  static String getDisplayName(String type) {
    switch (type) {
      case note:
        return 'Note';
      case task:
        return 'Task';
      case project:
        return 'Project';
      default:
        return type;
    }
  }
}