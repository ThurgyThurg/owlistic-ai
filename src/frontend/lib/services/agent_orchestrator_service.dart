import 'dart:convert';
import 'package:owlistic/services/base_service.dart';
import 'package:owlistic/utils/logger.dart';

/// Service for managing AI agent orchestration and chains
class AgentOrchestratorService extends BaseService {
  final Logger _logger = Logger('AgentOrchestratorService');

  /// Execute an agent chain
  Future<ChainExecutionResult?> executeChain(ChainExecutionRequest request) async {
    try {
      final response = await authenticatedPost(
        '/api/v1/agents/orchestrator/chains/execute',
        request.toJson(),
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return ChainExecutionResult.fromJson(data['execution']);
      } else {
        _logger.error('Failed to execute chain: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error executing chain', e);
      return null;
    }
  }

  /// Get all active executions
  Future<List<ChainExecutionResult>> getActiveExecutions() async {
    try {
      final response = await authenticatedGet('/api/v1/agents/orchestrator/executions');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final executions = data['executions'] as Map<String, dynamic>;
        return executions.values
            .map((e) => ChainExecutionResult.fromJson(e))
            .toList();
      } else {
        _logger.error('Failed to get executions: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error getting executions', e);
      return [];
    }
  }

  /// Get execution status by ID
  Future<ChainExecutionResult?> getExecutionStatus(String executionId) async {
    try {
      final response = await authenticatedGet(
        '/api/v1/agents/orchestrator/executions/$executionId',
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return ChainExecutionResult.fromJson(data['execution']);
      } else if (response.statusCode == 404) {
        _logger.warning('Execution not found: $executionId');
        return null;
      } else {
        _logger.error('Failed to get execution status: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error getting execution status', e);
      return null;
    }
  }

  /// List all available chains
  Future<List<AgentChain>> listChains() async {
    try {
      final response = await authenticatedGet('/api/v1/agents/orchestrator/chains');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final chains = data['chains'] as List<dynamic>;
        return chains.map((c) => AgentChain.fromJson(c)).toList();
      } else {
        _logger.error('Failed to list chains: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error listing chains', e);
      return [];
    }
  }

  /// Get chain details
  Future<AgentChain?> getChain(String chainId) async {
    try {
      final response = await authenticatedGet(
        '/api/v1/agents/orchestrator/chains/$chainId',
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return AgentChain.fromJson(data['chain']);
      } else if (response.statusCode == 404) {
        _logger.warning('Chain not found: $chainId');
        return null;
      } else {
        _logger.error('Failed to get chain: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error getting chain', e);
      return null;
    }
  }

  /// Create a new chain
  Future<AgentChain?> createChain(AgentChain chain) async {
    try {
      final response = await authenticatedPost(
        '/api/v1/agents/orchestrator/chains',
        chain.toJson(),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return AgentChain.fromJson(data['chain']);
      } else {
        _logger.error('Failed to create chain: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error creating chain', e);
      return null;
    }
  }

  /// Update an existing chain
  Future<AgentChain?> updateChain(String chainId, AgentChain chain) async {
    try {
      final response = await authenticatedPut(
        '/api/v1/agents/orchestrator/chains/$chainId',
        chain.toJson(),
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return AgentChain.fromJson(data['chain']);
      } else {
        _logger.error('Failed to update chain: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error updating chain', e);
      return null;
    }
  }

  /// Delete a chain
  Future<bool> deleteChain(String chainId) async {
    try {
      final response = await authenticatedDelete(
        '/api/v1/agents/orchestrator/chains/$chainId',
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        _logger.error('Failed to delete chain: ${response.statusCode} ${response.body}');
        return false;
      }
    } catch (e) {
      _logger.error('Error deleting chain', e);
      return false;
    }
  }

  /// Get available agent types
  Future<List<AgentType>> getAgentTypes() async {
    try {
      final response = await authenticatedGet(
        '/api/v1/agents/orchestrator/agent-types',
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final types = data['agent_types'] as List<dynamic>;
        return types.map((t) => AgentType.fromJson(t)).toList();
      } else {
        _logger.error('Failed to get agent types: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error getting agent types', e);
      return [];
    }
  }

  /// Get chain templates
  Future<List<ChainTemplate>> getChainTemplates() async {
    try {
      final response = await authenticatedGet(
        '/api/v1/agents/orchestrator/templates',
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final templates = data['templates'] as List<dynamic>;
        return templates.map((t) => ChainTemplate.fromJson(t)).toList();
      } else {
        _logger.error('Failed to get templates: ${response.statusCode} ${response.body}');
        return [];
      }
    } catch (e) {
      _logger.error('Error getting templates', e);
      return [];
    }
  }

  /// Instantiate a chain from a template
  Future<AgentChain?> instantiateTemplate(
    String templateId,
    Map<String, dynamic> parameters,
  ) async {
    try {
      final response = await authenticatedPost(
        '/api/v1/agents/orchestrator/templates/$templateId/instantiate',
        parameters,
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return AgentChain.fromJson(data['chain']);
      } else {
        _logger.error('Failed to instantiate template: ${response.statusCode} ${response.body}');
        return null;
      }
    } catch (e) {
      _logger.error('Error instantiating template', e);
      return null;
    }
  }
}

// Data models for agent orchestration

class ChainExecutionRequest {
  final String chainId;
  final Map<String, dynamic> initialData;

  ChainExecutionRequest({
    required this.chainId,
    required this.initialData,
  });

  Map<String, dynamic> toJson() => {
        'chain_id': chainId,
        'initial_data': initialData,
      };
}

class ChainExecutionResult {
  final String id;
  final String chainId;
  final String status;
  final DateTime startTime;
  final DateTime? endTime;
  final Map<String, dynamic> results;
  final List<AgentExecutionError> errors;
  final List<AgentExecutionLog> executionLog;

  ChainExecutionResult({
    required this.id,
    required this.chainId,
    required this.status,
    required this.startTime,
    this.endTime,
    required this.results,
    required this.errors,
    required this.executionLog,
  });

  factory ChainExecutionResult.fromJson(Map<String, dynamic> json) {
    return ChainExecutionResult(
      id: json['id'],
      chainId: json['chain_id'],
      status: json['status'],
      startTime: DateTime.parse(json['start_time']),
      endTime: json['end_time'] != null ? DateTime.parse(json['end_time']) : null,
      results: json['results'] ?? {},
      errors: (json['errors'] as List? ?? [])
          .map((e) => AgentExecutionError.fromJson(e))
          .toList(),
      executionLog: (json['execution_log'] as List? ?? [])
          .map((e) => AgentExecutionLog.fromJson(e))
          .toList(),
    );
  }
}

class AgentExecutionError {
  final String agentId;
  final String agentName;
  final String error;
  final DateTime timestamp;

  AgentExecutionError({
    required this.agentId,
    required this.agentName,
    required this.error,
    required this.timestamp,
  });

  factory AgentExecutionError.fromJson(Map<String, dynamic> json) {
    return AgentExecutionError(
      agentId: json['agent_id'],
      agentName: json['agent_name'],
      error: json['error'],
      timestamp: DateTime.parse(json['timestamp']),
    );
  }
}

class AgentExecutionLog {
  final String agentId;
  final String agentName;
  final String status;
  final Map<String, dynamic> input;
  final dynamic output;
  final DateTime startTime;
  final DateTime endTime;
  final double durationSeconds;

  AgentExecutionLog({
    required this.agentId,
    required this.agentName,
    required this.status,
    required this.input,
    required this.output,
    required this.startTime,
    required this.endTime,
    required this.durationSeconds,
  });

  factory AgentExecutionLog.fromJson(Map<String, dynamic> json) {
    return AgentExecutionLog(
      agentId: json['agent_id'],
      agentName: json['agent_name'],
      status: json['status'],
      input: json['input'] ?? {},
      output: json['output'],
      startTime: DateTime.parse(json['start_time']),
      endTime: DateTime.parse(json['end_time']),
      durationSeconds: json['duration_seconds'].toDouble(),
    );
  }
}

class AgentChain {
  final String? id;
  final String name;
  final String description;
  final String mode;
  final List<AgentDefinition> agents;
  final int timeoutSeconds;

  AgentChain({
    this.id,
    required this.name,
    required this.description,
    required this.mode,
    required this.agents,
    this.timeoutSeconds = 300,
  });

  factory AgentChain.fromJson(Map<String, dynamic> json) {
    return AgentChain(
      id: json['id'],
      name: json['name'],
      description: json['description'],
      mode: json['mode'],
      agents: (json['agents'] as List? ?? [])
          .map((a) => AgentDefinition.fromJson(a))
          .toList(),
      timeoutSeconds: json['timeout_seconds'] ?? 300,
    );
  }

  Map<String, dynamic> toJson() => {
        if (id != null) 'id': id,
        'name': name,
        'description': description,
        'mode': mode,
        'agents': agents.map((a) => a.toJson()).toList(),
        'timeout_seconds': timeoutSeconds,
      };
}

class AgentDefinition {
  final String? id;
  final String type;
  final String name;
  final String description;
  final Map<String, dynamic> config;
  final Map<String, String> inputMapping;
  final String outputKey;
  final List<ChainCondition> conditions;
  final RetryPolicy? retryPolicy;

  AgentDefinition({
    this.id,
    required this.type,
    required this.name,
    required this.description,
    this.config = const {},
    this.inputMapping = const {},
    required this.outputKey,
    this.conditions = const [],
    this.retryPolicy,
  });

  factory AgentDefinition.fromJson(Map<String, dynamic> json) {
    return AgentDefinition(
      id: json['id'],
      type: json['type'],
      name: json['name'],
      description: json['description'],
      config: json['config'] ?? {},
      inputMapping: Map<String, String>.from(json['input_mapping'] ?? {}),
      outputKey: json['output_key'] ?? '',
      conditions: (json['conditions'] as List? ?? [])
          .map((c) => ChainCondition.fromJson(c))
          .toList(),
      retryPolicy: json['retry_policy'] != null
          ? RetryPolicy.fromJson(json['retry_policy'])
          : null,
    );
  }

  Map<String, dynamic> toJson() => {
        if (id != null) 'id': id,
        'type': type,
        'name': name,
        'description': description,
        'config': config,
        'input_mapping': inputMapping,
        'output_key': outputKey,
        'conditions': conditions.map((c) => c.toJson()).toList(),
        if (retryPolicy != null) 'retry_policy': retryPolicy!.toJson(),
      };
}

class ChainCondition {
  final String type;
  final String dataKey;
  final dynamic value;
  final String? operator;

  ChainCondition({
    required this.type,
    required this.dataKey,
    required this.value,
    this.operator,
  });

  factory ChainCondition.fromJson(Map<String, dynamic> json) {
    return ChainCondition(
      type: json['type'],
      dataKey: json['data_key'],
      value: json['value'],
      operator: json['operator'],
    );
  }

  Map<String, dynamic> toJson() => {
        'type': type,
        'data_key': dataKey,
        'value': value,
        if (operator != null) 'operator': operator,
      };
}

class RetryPolicy {
  final int maxRetries;
  final int backoffSeconds;
  final List<String> retryOnErrors;

  RetryPolicy({
    this.maxRetries = 3,
    this.backoffSeconds = 1,
    this.retryOnErrors = const [],
  });

  factory RetryPolicy.fromJson(Map<String, dynamic> json) {
    return RetryPolicy(
      maxRetries: json['max_retries'] ?? 3,
      backoffSeconds: json['backoff_seconds'] ?? 1,
      retryOnErrors: List<String>.from(json['retry_on_errors'] ?? []),
    );
  }

  Map<String, dynamic> toJson() => {
        'max_retries': maxRetries,
        'backoff_seconds': backoffSeconds,
        'retry_on_errors': retryOnErrors,
      };
}

class AgentType {
  final String type;
  final String name;
  final String description;
  final Map<String, String> inputSchema;

  AgentType({
    required this.type,
    required this.name,
    required this.description,
    required this.inputSchema,
  });

  factory AgentType.fromJson(Map<String, dynamic> json) {
    return AgentType(
      type: json['type'],
      name: json['name'],
      description: json['description'],
      inputSchema: Map<String, String>.from(json['input_schema'] ?? {}),
    );
  }
}

class ChainTemplate {
  final String id;
  final String name;
  final String description;
  final List<TemplateParameter> parameters;

  ChainTemplate({
    required this.id,
    required this.name,
    required this.description,
    required this.parameters,
  });

  factory ChainTemplate.fromJson(Map<String, dynamic> json) {
    return ChainTemplate(
      id: json['id'],
      name: json['name'],
      description: json['description'],
      parameters: (json['parameters'] as List? ?? [])
          .map((p) => TemplateParameter.fromJson(p))
          .toList(),
    );
  }
}

class TemplateParameter {
  final String name;
  final String type;
  final String description;

  TemplateParameter({
    required this.name,
    required this.type,
    required this.description,
  });

  factory TemplateParameter.fromJson(Map<String, dynamic> json) {
    return TemplateParameter(
      name: json['name'],
      type: json['type'],
      description: json['description'],
    );
  }
}