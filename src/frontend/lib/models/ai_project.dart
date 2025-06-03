class AIProject {
  final String id;
  final String userId;
  final String name;
  final String? description;
  final String status;
  final String? notebookId;
  final List<String>? aiTags;
  final Map<String, dynamic>? aiMetadata;
  final List<String>? relatedNoteIds;
  final DateTime createdAt;
  final DateTime updatedAt;

  const AIProject({
    required this.id,
    required this.userId,
    required this.name,
    this.description,
    required this.status,
    this.notebookId,
    this.aiTags,
    this.aiMetadata,
    this.relatedNoteIds,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AIProject.fromJson(Map<String, dynamic> json) {
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

    List<String>? aiTags;
    if (json['ai_tags'] != null && json['ai_tags'] is List) {
      aiTags = (json['ai_tags'] as List).cast<String>();
    }

    Map<String, dynamic>? aiMetadata;
    if (json['ai_metadata'] != null && json['ai_metadata'] is Map) {
      aiMetadata = Map<String, dynamic>.from(json['ai_metadata']);
    }

    List<String>? relatedNoteIds;
    if (json['related_note_ids'] != null && json['related_note_ids'] is List) {
      relatedNoteIds = (json['related_note_ids'] as List).cast<String>();
    }

    return AIProject(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      name: json['name'] ?? '',
      description: json['description'],
      status: json['status'] ?? '',
      notebookId: json['notebook_id'],
      aiTags: aiTags,
      aiMetadata: aiMetadata,
      relatedNoteIds: relatedNoteIds,
      createdAt: createdAt,
      updatedAt: updatedAt,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'name': name,
      if (description != null) 'description': description,
      'status': status,
      if (notebookId != null) 'notebook_id': notebookId,
      if (aiTags != null) 'ai_tags': aiTags,
      if (aiMetadata != null) 'ai_metadata': aiMetadata,
      if (relatedNoteIds != null) 'related_note_ids': relatedNoteIds,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  AIProject copyWith({
    String? id,
    String? userId,
    String? name,
    String? description,
    String? status,
    String? notebookId,
    List<String>? aiTags,
    Map<String, dynamic>? aiMetadata,
    List<String>? relatedNoteIds,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return AIProject(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      name: name ?? this.name,
      description: description ?? this.description,
      status: status ?? this.status,
      notebookId: notebookId ?? this.notebookId,
      aiTags: aiTags ?? this.aiTags,
      aiMetadata: aiMetadata ?? this.aiMetadata,
      relatedNoteIds: relatedNoteIds ?? this.relatedNoteIds,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  bool get isActive => status == 'active';
  bool get isCompleted => status == 'completed';
  bool get isArchived => status == 'archived';
  bool get isPaused => status == 'paused';
  bool get hasNotebook => notebookId != null;
  bool get hasNotes => relatedNoteIds != null && relatedNoteIds!.isNotEmpty;
}

class TaskBreakdownRequest {
  final String goal;
  final String? context;
  final String? timeFrame;
  final int? maxSteps;
  final String? priority;
  final Map<String, dynamic>? preferences;

  const TaskBreakdownRequest({
    required this.goal,
    this.context,
    this.timeFrame,
    this.maxSteps,
    this.priority,
    this.preferences,
  });

  Map<String, dynamic> toJson() {
    return {
      'goal': goal,
      if (context != null) 'context': context,
      if (timeFrame != null) 'time_frame': timeFrame,
      if (maxSteps != null) 'max_steps': maxSteps,
      if (priority != null) 'priority': priority,
      if (preferences != null) 'preferences': preferences,
    };
  }
}

class TaskBreakdownResponse {
  final String goal;
  final List<TaskStep> steps;
  final String? estimatedTimeframe;
  final String? complexity;
  final List<String>? prerequisites;
  final List<String>? resources;
  final Map<String, dynamic>? metadata;

  const TaskBreakdownResponse({
    required this.goal,
    required this.steps,
    this.estimatedTimeframe,
    this.complexity,
    this.prerequisites,
    this.resources,
    this.metadata,
  });

  factory TaskBreakdownResponse.fromJson(Map<String, dynamic> json) {
    List<TaskStep> steps = [];
    if (json['steps'] != null && json['steps'] is List) {
      steps = (json['steps'] as List).asMap().entries.map((entry) {
        final index = entry.key;
        final step = entry.value;
        
        // Handle both simple string format (from backend) and object format
        if (step is String) {
          return TaskStep(
            stepNumber: index + 1,
            title: 'Step ${index + 1}',
            description: step,
          );
        } else if (step is Map<String, dynamic>) {
          return TaskStep.fromJson(step);
        } else {
          // Fallback for unexpected formats
          return TaskStep(
            stepNumber: index + 1,
            title: 'Step ${index + 1}',
            description: step.toString(),
          );
        }
      }).toList();
    }

    List<String>? prerequisites;
    if (json['prerequisites'] != null && json['prerequisites'] is List) {
      prerequisites = (json['prerequisites'] as List).cast<String>();
    }

    List<String>? resources;
    if (json['resources'] != null && json['resources'] is List) {
      resources = (json['resources'] as List).cast<String>();
    }

    Map<String, dynamic>? metadata;
    if (json['metadata'] != null && json['metadata'] is Map) {
      metadata = Map<String, dynamic>.from(json['metadata']);
    }

    return TaskBreakdownResponse(
      goal: json['goal'] ?? '',
      steps: steps,
      estimatedTimeframe: json['estimated_timeframe'],
      complexity: json['complexity'],
      prerequisites: prerequisites,
      resources: resources,
      metadata: metadata,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'goal': goal,
      'steps': steps.map((step) => step.toJson()).toList(),
      if (estimatedTimeframe != null) 'estimated_timeframe': estimatedTimeframe,
      if (complexity != null) 'complexity': complexity,
      if (prerequisites != null) 'prerequisites': prerequisites,
      if (resources != null) 'resources': resources,
      if (metadata != null) 'metadata': metadata,
    };
  }
}

class TaskStep {
  final int stepNumber;
  final String title;
  final String description;
  final String? estimatedDuration;
  final String? difficulty;
  final List<String>? dependencies;
  final List<String>? deliverables;
  final Map<String, dynamic>? metadata;
  final bool isCompleted;
  final DateTime? scheduledStart;
  final DateTime? scheduledEnd;
  final DateTime? actualStart;
  final DateTime? actualEnd;

  const TaskStep({
    required this.stepNumber,
    required this.title,
    required this.description,
    this.estimatedDuration,
    this.difficulty,
    this.dependencies,
    this.deliverables,
    this.metadata,
    this.isCompleted = false,
    this.scheduledStart,
    this.scheduledEnd,
    this.actualStart,
    this.actualEnd,
  });

  factory TaskStep.fromJson(Map<String, dynamic> json) {
    List<String>? dependencies;
    if (json['dependencies'] != null && json['dependencies'] is List) {
      dependencies = (json['dependencies'] as List).cast<String>();
    }

    List<String>? deliverables;
    if (json['deliverables'] != null && json['deliverables'] is List) {
      deliverables = (json['deliverables'] as List).cast<String>();
    }

    Map<String, dynamic>? metadata;
    if (json['metadata'] != null && json['metadata'] is Map) {
      metadata = Map<String, dynamic>.from(json['metadata']);
    }

    DateTime? scheduledStart;
    if (json['scheduled_start'] != null) {
      try {
        scheduledStart = DateTime.parse(json['scheduled_start']);
      } catch (e) {
        scheduledStart = null;
      }
    }

    DateTime? scheduledEnd;
    if (json['scheduled_end'] != null) {
      try {
        scheduledEnd = DateTime.parse(json['scheduled_end']);
      } catch (e) {
        scheduledEnd = null;
      }
    }

    DateTime? actualStart;
    if (json['actual_start'] != null) {
      try {
        actualStart = DateTime.parse(json['actual_start']);
      } catch (e) {
        actualStart = null;
      }
    }

    DateTime? actualEnd;
    if (json['actual_end'] != null) {
      try {
        actualEnd = DateTime.parse(json['actual_end']);
      } catch (e) {
        actualEnd = null;
      }
    }

    return TaskStep(
      stepNumber: json['step_number'] ?? 0,
      title: json['title'] ?? '',
      description: json['description'] ?? '',
      estimatedDuration: json['estimated_duration'],
      difficulty: json['difficulty'],
      dependencies: dependencies,
      deliverables: deliverables,
      metadata: metadata,
      isCompleted: json['is_completed'] ?? false,
      scheduledStart: scheduledStart,
      scheduledEnd: scheduledEnd,
      actualStart: actualStart,
      actualEnd: actualEnd,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'step_number': stepNumber,
      'title': title,
      'description': description,
      if (estimatedDuration != null) 'estimated_duration': estimatedDuration,
      if (difficulty != null) 'difficulty': difficulty,
      if (dependencies != null) 'dependencies': dependencies,
      if (deliverables != null) 'deliverables': deliverables,
      if (metadata != null) 'metadata': metadata,
      'is_completed': isCompleted,
      if (scheduledStart != null) 'scheduled_start': scheduledStart!.toIso8601String(),
      if (scheduledEnd != null) 'scheduled_end': scheduledEnd!.toIso8601String(),
      if (actualStart != null) 'actual_start': actualStart!.toIso8601String(),
      if (actualEnd != null) 'actual_end': actualEnd!.toIso8601String(),
    };
  }

  TaskStep copyWith({
    int? stepNumber,
    String? title,
    String? description,
    String? estimatedDuration,
    String? difficulty,
    List<String>? dependencies,
    List<String>? deliverables,
    Map<String, dynamic>? metadata,
    bool? isCompleted,
    DateTime? scheduledStart,
    DateTime? scheduledEnd,
    DateTime? actualStart,
    DateTime? actualEnd,
  }) {
    return TaskStep(
      stepNumber: stepNumber ?? this.stepNumber,
      title: title ?? this.title,
      description: description ?? this.description,
      estimatedDuration: estimatedDuration ?? this.estimatedDuration,
      difficulty: difficulty ?? this.difficulty,
      dependencies: dependencies ?? this.dependencies,
      deliverables: deliverables ?? this.deliverables,
      metadata: metadata ?? this.metadata,
      isCompleted: isCompleted ?? this.isCompleted,
      scheduledStart: scheduledStart ?? this.scheduledStart,
      scheduledEnd: scheduledEnd ?? this.scheduledEnd,
      actualStart: actualStart ?? this.actualStart,
      actualEnd: actualEnd ?? this.actualEnd,
    );
  }

  bool get isScheduled => scheduledStart != null && scheduledEnd != null;
  bool get isInProgress => actualStart != null && actualEnd == null;
  bool get isOverdue => scheduledEnd != null && 
                       !isCompleted && 
                       DateTime.now().isAfter(scheduledEnd!);
}