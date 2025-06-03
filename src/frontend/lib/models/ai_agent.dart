class AIAgent {
  final String id;
  final String userId;
  final String agentType;
  final String status;
  final Map<String, dynamic>? inputData;
  final Map<String, dynamic>? outputData;
  final List<AIAgentStep>? steps;
  final String? error;
  final DateTime? startedAt;
  final DateTime? completedAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  const AIAgent({
    required this.id,
    required this.userId,
    required this.agentType,
    required this.status,
    this.inputData,
    this.outputData,
    this.steps,
    this.error,
    this.startedAt,
    this.completedAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AIAgent.fromJson(Map<String, dynamic> json) {
    DateTime? startedAt;
    if (json['started_at'] != null) {
      try {
        startedAt = DateTime.parse(json['started_at']);
      } catch (e) {
        startedAt = null;
      }
    }

    DateTime? completedAt;
    if (json['completed_at'] != null) {
      try {
        completedAt = DateTime.parse(json['completed_at']);
      } catch (e) {
        completedAt = null;
      }
    }

    DateTime createdAt;
    try {
      createdAt = DateTime.parse(json['created_at']);
    } catch (e) {
      createdAt = DateTime.now();
    }

    DateTime updatedAt;
    try {
      updatedAt = DateTime.parse(json['updated_at']);
    } catch (e) {
      updatedAt = DateTime.now();
    }

    List<AIAgentStep>? steps;
    if (json['steps'] != null && json['steps'] is List) {
      steps = (json['steps'] as List)
          .map((step) => AIAgentStep.fromJson(step))
          .toList();
    }

    Map<String, dynamic>? inputData;
    if (json['input_data'] != null && json['input_data'] is Map) {
      inputData = Map<String, dynamic>.from(json['input_data']);
    }

    Map<String, dynamic>? outputData;
    if (json['output_data'] != null && json['output_data'] is Map) {
      outputData = Map<String, dynamic>.from(json['output_data']);
    }

    return AIAgent(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      agentType: json['agent_type'] ?? '',
      status: json['status'] ?? '',
      inputData: inputData,
      outputData: outputData,
      steps: steps,
      error: json['error'],
      startedAt: startedAt,
      completedAt: completedAt,
      createdAt: createdAt,
      updatedAt: updatedAt,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'agent_type': agentType,
      'status': status,
      if (inputData != null) 'input_data': inputData,
      if (outputData != null) 'output_data': outputData,
      if (steps != null) 'steps': steps?.map((step) => step.toJson()).toList(),
      if (error != null) 'error': error,
      if (startedAt != null) 'started_at': startedAt!.toIso8601String(),
      if (completedAt != null) 'completed_at': completedAt!.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  AIAgent copyWith({
    String? id,
    String? userId,
    String? agentType,
    String? status,
    Map<String, dynamic>? inputData,
    Map<String, dynamic>? outputData,
    List<AIAgentStep>? steps,
    String? error,
    DateTime? startedAt,
    DateTime? completedAt,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return AIAgent(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      agentType: agentType ?? this.agentType,
      status: status ?? this.status,
      inputData: inputData ?? this.inputData,
      outputData: outputData ?? this.outputData,
      steps: steps ?? this.steps,
      error: error ?? this.error,
      startedAt: startedAt ?? this.startedAt,
      completedAt: completedAt ?? this.completedAt,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  bool get isRunning => status == 'running';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isPending => status == 'pending';

  Duration? get executionTime {
    if (startedAt != null && completedAt != null) {
      return completedAt!.difference(startedAt!);
    }
    return null;
  }

  int get completedStepsCount {
    if (steps == null) return 0;
    return steps!.where((step) => step.status == 'completed').length;
  }

  int get totalStepsCount => steps?.length ?? 0;

  double get progressPercentage {
    if (totalStepsCount == 0) return 0.0;
    return (completedStepsCount / totalStepsCount) * 100;
  }
}

class AIAgentStep {
  final String id;
  final String agentId;
  final int stepNumber;
  final String name;
  final String? description;
  final String status;
  final Map<String, dynamic>? inputData;
  final Map<String, dynamic>? outputData;
  final String? error;
  final DateTime? startedAt;
  final DateTime? completedAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  const AIAgentStep({
    required this.id,
    required this.agentId,
    required this.stepNumber,
    required this.name,
    this.description,
    required this.status,
    this.inputData,
    this.outputData,
    this.error,
    this.startedAt,
    this.completedAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AIAgentStep.fromJson(Map<String, dynamic> json) {
    DateTime? startedAt;
    if (json['started_at'] != null) {
      try {
        startedAt = DateTime.parse(json['started_at']);
      } catch (e) {
        startedAt = null;
      }
    }

    DateTime? completedAt;
    if (json['completed_at'] != null) {
      try {
        completedAt = DateTime.parse(json['completed_at']);
      } catch (e) {
        completedAt = null;
      }
    }

    DateTime createdAt;
    try {
      createdAt = DateTime.parse(json['created_at']);
    } catch (e) {
      createdAt = DateTime.now();
    }

    DateTime updatedAt;
    try {
      updatedAt = DateTime.parse(json['updated_at']);
    } catch (e) {
      updatedAt = DateTime.now();
    }

    Map<String, dynamic>? inputData;
    if (json['input_data'] != null && json['input_data'] is Map) {
      inputData = Map<String, dynamic>.from(json['input_data']);
    }

    Map<String, dynamic>? outputData;
    if (json['output_data'] != null && json['output_data'] is Map) {
      outputData = Map<String, dynamic>.from(json['output_data']);
    }

    return AIAgentStep(
      id: json['id'] ?? '',
      agentId: json['agent_id'] ?? '',
      stepNumber: json['step_number'] ?? 0,
      name: json['name'] ?? '',
      description: json['description'],
      status: json['status'] ?? '',
      inputData: inputData,
      outputData: outputData,
      error: json['error'],
      startedAt: startedAt,
      completedAt: completedAt,
      createdAt: createdAt,
      updatedAt: updatedAt,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'agent_id': agentId,
      'step_number': stepNumber,
      'name': name,
      if (description != null) 'description': description,
      'status': status,
      if (inputData != null) 'input_data': inputData,
      if (outputData != null) 'output_data': outputData,
      if (error != null) 'error': error,
      if (startedAt != null) 'started_at': startedAt!.toIso8601String(),
      if (completedAt != null) 'completed_at': completedAt!.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  AIAgentStep copyWith({
    String? id,
    String? agentId,
    int? stepNumber,
    String? name,
    String? description,
    String? status,
    Map<String, dynamic>? inputData,
    Map<String, dynamic>? outputData,
    String? error,
    DateTime? startedAt,
    DateTime? completedAt,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return AIAgentStep(
      id: id ?? this.id,
      agentId: agentId ?? this.agentId,
      stepNumber: stepNumber ?? this.stepNumber,
      name: name ?? this.name,
      description: description ?? this.description,
      status: status ?? this.status,
      inputData: inputData ?? this.inputData,
      outputData: outputData ?? this.outputData,
      error: error ?? this.error,
      startedAt: startedAt ?? this.startedAt,
      completedAt: completedAt ?? this.completedAt,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  bool get isRunning => status == 'running';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isPending => status == 'pending';

  Duration? get executionTime {
    if (startedAt != null && completedAt != null) {
      return completedAt!.difference(startedAt!);
    }
    return null;
  }
}