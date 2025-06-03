import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../models/ai_project.dart';
import '../models/ai_agent.dart';
import '../services/ai_service.dart';
import '../widgets/app_bar_common.dart';
import '../utils/logger.dart';

class AIDashboardScreen extends StatefulWidget {
  const AIDashboardScreen({Key? key}) : super(key: key);

  @override
  State<AIDashboardScreen> createState() => _AIDashboardScreenState();
}

class _AIDashboardScreenState extends State<AIDashboardScreen>
    with SingleTickerProviderStateMixin {
  final Logger _logger = Logger('AIDashboardScreen');
  final TextEditingController _goalController = TextEditingController();
  final TextEditingController _contextController = TextEditingController();
  
  TaskBreakdownResponse? _currentBreakdown;
  bool _isBreakingDown = false;
  String? _error;
  late TabController _tabController;

  List<AIAgent> _recentAgents = [];
  List<AIProject> _projects = [];
  bool _isLoadingAgents = false;
  bool _isLoadingProjects = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadRecentAgents();
    _loadProjects();
  }

  @override
  void dispose() {
    _goalController.dispose();
    _contextController.dispose();
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadRecentAgents() async {
    setState(() {
      _isLoadingAgents = true;
    });

    try {
      final response = await AIService().getAgentRuns();
      if (response['status'] == 'success') {
        setState(() {
          _recentAgents = (response['data'] as List)
              .map((agent) => AIAgent.fromJson(agent))
              .toList();
        });
      }
    } catch (e) {
      _logger.error('Failed to load recent agents: $e');
    } finally {
      setState(() {
        _isLoadingAgents = false;
      });
    }
  }

  Future<void> _loadProjects() async {
    setState(() {
      _isLoadingProjects = true;
    });

    try {
      final response = await AIService().getAIProjects();
      if (response['status'] == 'success') {
        setState(() {
          _projects = (response['data'] as List)
              .map((project) => AIProject.fromJson(project))
              .toList();
        });
      }
    } catch (e) {
      _logger.error('Failed to load projects: $e');
    } finally {
      setState(() {
        _isLoadingProjects = false;
      });
    }
  }

  Future<void> _breakDownTask() async {
    if (_goalController.text.trim().isEmpty) {
      setState(() {
        _error = 'Please enter a goal to break down';
      });
      return;
    }

    setState(() {
      _isBreakingDown = true;
      _error = null;
    });

    try {
      final request = TaskBreakdownRequest(
        goal: _goalController.text.trim(),
        context: _contextController.text.trim().isEmpty 
            ? null 
            : _contextController.text.trim(),
        maxSteps: 10,
        priority: 'medium',
      );

      // This will call the backend task breakdown service
      final response = await AIService().breakDownTask(request);
      
      setState(() {
        _currentBreakdown = TaskBreakdownResponse.fromJson(response['data']);
        _isBreakingDown = false;
      });

      _logger.info('Task breakdown completed successfully');
    } catch (e) {
      _logger.error('Task breakdown failed: $e');
      setState(() {
        _error = 'Failed to break down task: ${e.toString()}';
        _isBreakingDown = false;
      });
    }
  }

  Future<void> _createProject() async {
    if (_currentBreakdown == null) return;

    try {
      final metadata = {
        'breakdown': _currentBreakdown!.toJson(),
        'created_from': 'task_breakdown',
        'step_count': _currentBreakdown!.steps.length,
      };

      final response = await AIService().createAIProject(
        name: _currentBreakdown!.goal,
        description: 'Project created from AI task breakdown',
        aiTags: ['task-breakdown', 'ai-generated'],
        aiMetadata: metadata,
      );

      if (response['status'] == 'success') {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Project created successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        _loadProjects();
      }
    } catch (e) {
      _logger.error('Failed to create project: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to create project: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBarCommon(
        title: 'AI Dashboard',
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              _loadRecentAgents();
              _loadProjects();
            },
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: Column(
        children: [
          TabBar(
            controller: _tabController,
            tabs: const [
              Tab(text: 'Task Breakdown', icon: Icon(Icons.task_alt)),
              Tab(text: 'Agents', icon: Icon(Icons.smart_toy)),
              Tab(text: 'Projects', icon: Icon(Icons.folder_special)),
            ],
          ),
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildTaskBreakdownTab(),
                _buildAgentsTab(),
                _buildProjectsTab(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTaskBreakdownTab() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.psychology,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'AI Task Breakdown',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _goalController,
                    decoration: const InputDecoration(
                      labelText: 'Goal or Task',
                      hintText: 'e.g., "Learn Flutter development" or "Build a mobile app"',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 2,
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _contextController,
                    decoration: const InputDecoration(
                      labelText: 'Context (Optional)',
                      hintText: 'Additional details, constraints, or preferences...',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 3,
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: ElevatedButton.icon(
                          onPressed: _isBreakingDown ? null : _breakDownTask,
                          icon: _isBreakingDown
                              ? const SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(strokeWidth: 2),
                                )
                              : const Icon(Icons.auto_fix_high),
                          label: Text(
                            _isBreakingDown ? 'Breaking Down...' : 'Break Down Task',
                          ),
                        ),
                      ),
                      if (_currentBreakdown != null) ...[
                        const SizedBox(width: 12),
                        ElevatedButton.icon(
                          onPressed: _createProject,
                          icon: const Icon(Icons.add),
                          label: const Text('Create Project'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Theme.of(context).colorScheme.secondary,
                          ),
                        ),
                      ],
                    ],
                  ),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.errorContainer,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            Icons.error,
                            color: Theme.of(context).colorScheme.onErrorContainer,
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              _error!,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.onErrorContainer,
                              ),
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
          if (_currentBreakdown != null) ...[
            const SizedBox(height: 16),
            _buildBreakdownResults(),
          ],
        ],
      ),
    );
  }

  Widget _buildBreakdownResults() {
    if (_currentBreakdown == null) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.list_alt,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    _currentBreakdown!.goal,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                Chip(
                  label: Text('${_currentBreakdown!.steps.length} steps'),
                  backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                ),
              ],
            ),
            if (_currentBreakdown!.estimatedTimeframe != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  Icon(
                    Icons.schedule,
                    size: 16,
                    color: Theme.of(context).colorScheme.secondary,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Estimated time: ${_currentBreakdown!.estimatedTimeframe}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ],
            if (_currentBreakdown!.complexity != null) ...[
              const SizedBox(height: 4),
              Row(
                children: [
                  Icon(
                    Icons.trending_up,
                    size: 16,
                    color: Theme.of(context).colorScheme.secondary,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Complexity: ${_currentBreakdown!.complexity}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ],
            const SizedBox(height: 16),
            const Divider(),
            const SizedBox(height: 8),
            ...(_currentBreakdown!.steps.asMap().entries.map((entry) {
              final step = entry.value;
              return _buildStepCard(step);
            }).toList()),
            if (_currentBreakdown!.prerequisites?.isNotEmpty == true) ...[
              const SizedBox(height: 16),
              const Divider(),
              const SizedBox(height: 8),
              Text(
                'Prerequisites',
                style: Theme.of(context).textTheme.titleSmall,
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 4,
                children: _currentBreakdown!.prerequisites!
                    .map((prereq) => Chip(
                          label: Text(prereq),
                          backgroundColor: Theme.of(context).colorScheme.surfaceVariant,
                        ))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildStepCard(TaskStep step) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      child: Card(
        elevation: 1,
        child: Padding(
          padding: const EdgeInsets.all(12.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 28,
                    height: 28,
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.primary,
                      shape: BoxShape.circle,
                    ),
                    child: Center(
                      child: Text(
                        '${step.stepNumber}',
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onPrimary,
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      step.title,
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                  ),
                  if (step.estimatedDuration != null)
                    Chip(
                      label: Text(step.estimatedDuration!),
                      backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
                    ),
                ],
              ),
              const SizedBox(height: 8),
              Padding(
                padding: const EdgeInsets.only(left: 40),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      step.description,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    if (step.difficulty != null) ...[
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Icon(
                            Icons.signal_cellular_alt,
                            size: 16,
                            color: Theme.of(context).colorScheme.tertiary,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Difficulty: ${step.difficulty}',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ],
                      ),
                    ],
                    if (step.deliverables?.isNotEmpty == true) ...[
                      const SizedBox(height: 8),
                      Text(
                        'Deliverables:',
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                      const SizedBox(height: 4),
                      ...step.deliverables!.map((deliverable) => Padding(
                        padding: const EdgeInsets.only(left: 8),
                        child: Row(
                          children: [
                            const Icon(Icons.check_circle_outline, size: 12),
                            const SizedBox(width: 4),
                            Expanded(
                              child: Text(
                                deliverable,
                                style: Theme.of(context).textTheme.bodySmall,
                              ),
                            ),
                          ],
                        ),
                      )),
                    ],
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Spacer(),
                        TextButton.icon(
                          onPressed: () {
                            // TODO: Implement calendar scheduling
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Calendar integration coming soon!'),
                              ),
                            );
                          },
                          icon: const Icon(Icons.schedule, size: 16),
                          label: const Text('Schedule'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAgentsTab() {
    return RefreshIndicator(
      onRefresh: _loadRecentAgents,
      child: _isLoadingAgents
          ? const Center(child: CircularProgressIndicator())
          : _recentAgents.isEmpty
              ? const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.smart_toy, size: 64, color: Colors.grey),
                      SizedBox(height: 16),
                      Text('No AI agents yet'),
                      Text('Start by breaking down a task!'),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _recentAgents.length,
                  itemBuilder: (context, index) {
                    final agent = _recentAgents[index];
                    return _buildAgentCard(agent);
                  },
                ),
    );
  }

  Widget _buildAgentCard(AIAgent agent) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: agent.isCompleted 
              ? Colors.green 
              : agent.isRunning 
                  ? Colors.orange 
                  : Colors.grey,
          child: Icon(
            agent.isCompleted 
                ? Icons.check 
                : agent.isRunning 
                    ? Icons.hourglass_empty 
                    : Icons.error,
            color: Colors.white,
          ),
        ),
        title: Text(agent.agentType.replaceAll('_', ' ').toUpperCase()),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Status: ${agent.status}'),
            if (agent.totalStepsCount > 0)
              Text('Progress: ${agent.completedStepsCount}/${agent.totalStepsCount} steps'),
          ],
        ),
        trailing: Text(
          agent.createdAt.toLocal().toString().split(' ')[0],
          style: Theme.of(context).textTheme.bodySmall,
        ),
        onTap: () {
          // TODO: Navigate to agent details
        },
      ),
    );
  }

  Widget _buildProjectsTab() {
    return RefreshIndicator(
      onRefresh: _loadProjects,
      child: _isLoadingProjects
          ? const Center(child: CircularProgressIndicator())
          : _projects.isEmpty
              ? const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.folder_special, size: 64, color: Colors.grey),
                      SizedBox(height: 16),
                      Text('No AI projects yet'),
                      Text('Create one from a task breakdown!'),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _projects.length,
                  itemBuilder: (context, index) {
                    final project = _projects[index];
                    return _buildProjectCard(project);
                  },
                ),
    );
  }

  Widget _buildProjectCard(AIProject project) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Column(
        children: [
          ListTile(
            leading: CircleAvatar(
              backgroundColor: project.isActive 
                  ? Colors.blue 
                  : project.isCompleted 
                      ? Colors.green 
                      : Colors.grey,
              child: Icon(
                project.isActive 
                    ? Icons.play_arrow 
                    : project.isCompleted 
                        ? Icons.check 
                        : Icons.archive,
                color: Colors.white,
              ),
            ),
            title: Text(project.name),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (project.description != null) Text(project.description!),
                Text('Status: ${project.status}'),
                if (project.hasNotebook)
                  Row(
                    children: [
                      Icon(
                        Icons.folder_special,
                        size: 16,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'Notebook created',
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.primary,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                if (project.hasNotes)
                  Row(
                    children: [
                      Icon(
                        Icons.note,
                        size: 16,
                        color: Theme.of(context).colorScheme.secondary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${project.relatedNoteIds!.length} notes',
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.secondary,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                if (project.aiTags?.isNotEmpty == true) ...[
                  const SizedBox(height: 4),
                  Wrap(
                    spacing: 4,
                    children: project.aiTags!.take(3).map((tag) => Chip(
                      label: Text(tag),
                      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    )).toList(),
                  ),
                ],
              ],
            ),
            trailing: Text(
              project.createdAt.toLocal().toString().split(' ')[0],
              style: Theme.of(context).textTheme.bodySmall,
            ),
            onTap: () {
              // TODO: Navigate to project details
            },
          ),
          if (project.hasNotebook) ...[
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: () {
                        context.go('/notebooks/${project.notebookId}');
                      },
                      icon: const Icon(Icons.folder_open, size: 16),
                      label: const Text('Open Notebook'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                        foregroundColor: Theme.of(context).colorScheme.onPrimaryContainer,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  OutlinedButton.icon(
                    onPressed: () {
                      context.go('/notes');
                    },
                    icon: const Icon(Icons.list, size: 16),
                    label: const Text('View Notes'),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}