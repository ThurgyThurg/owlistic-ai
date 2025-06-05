import 'package:flutter/material.dart';
import 'package:owlistic/services/agent_orchestrator_service.dart';
import 'package:owlistic/widgets/app_bar_common.dart';
import 'package:owlistic/widgets/app_drawer.dart';
import 'package:owlistic/widgets/card_container.dart';
import 'package:owlistic/widgets/empty_state.dart';
import 'package:owlistic/utils/logger.dart';

class AgentOrchestratorScreen extends StatefulWidget {
  const AgentOrchestratorScreen({Key? key}) : super(key: key);

  @override
  State<AgentOrchestratorScreen> createState() => _AgentOrchestratorScreenState();
}

class _AgentOrchestratorScreenState extends State<AgentOrchestratorScreen> {
  final Logger _logger = Logger('AgentOrchestratorScreen');
  final AgentOrchestratorService _service = AgentOrchestratorService();
  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  List<AgentChain> _chains = [];
  List<ChainTemplate> _templates = [];
  List<ChainExecutionResult> _activeExecutions = [];
  bool _isLoading = true;
  int _selectedTab = 0;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    
    try {
      final chains = await _service.listChains();
      final templates = await _service.getChainTemplates();
      final executions = await _service.getActiveExecutions();
      
      if (mounted) {
        setState(() {
          _chains = chains;
          _templates = templates;
          _activeExecutions = executions;
          _isLoading = false;
        });
      }
    } catch (e) {
      _logger.error('Error loading data', e);
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to load data: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: _scaffoldKey,
      appBar: AppBarCommon(
        title: 'AI Agent Orchestrator',
        onMenuPressed: () => _scaffoldKey.currentState?.openDrawer(),
        showBackButton: false,
      ),
      drawer: const SidebarDrawer(),
      body: Column(
        children: [
          // Tab bar
          Container(
            color: Theme.of(context).colorScheme.surface,
            child: Row(
              children: [
                _buildTab('Chains', 0),
                _buildTab('Templates', 1),
                _buildTab('Executions', 2),
              ],
            ),
          ),
          // Content
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : IndexedStack(
                    index: _selectedTab,
                    children: [
                      _buildChainsTab(),
                      _buildTemplatesTab(),
                      _buildExecutionsTab(),
                    ],
                  ),
          ),
        ],
      ),
      floatingActionButton: _selectedTab == 0
          ? FloatingActionButton(
              onPressed: _showCreateChainDialog,
              child: const Icon(Icons.add),
              tooltip: 'Create new chain',
            )
          : null,
    );
  }

  Widget _buildTab(String title, int index) {
    final isSelected = _selectedTab == index;
    return Expanded(
      child: InkWell(
        onTap: () => setState(() => _selectedTab = index),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 16),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                color: isSelected
                    ? Theme.of(context).colorScheme.primary
                    : Colors.transparent,
                width: 2,
              ),
            ),
          ),
          child: Text(
            title,
            textAlign: TextAlign.center,
            style: TextStyle(
              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              color: isSelected
                  ? Theme.of(context).colorScheme.primary
                  : Theme.of(context).colorScheme.onSurface,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildChainsTab() {
    if (_chains.isEmpty) {
      return EmptyState(
        title: 'No Agent Chains',
        message: 'Create your first agent chain to automate complex workflows',
        icon: Icons.account_tree,
        actionLabel: 'Create Chain',
        onAction: _showCreateChainDialog,
      );
    }

    return RefreshIndicator(
      onRefresh: _loadData,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _chains.length,
        itemBuilder: (context, index) {
          final chain = _chains[index];
          return CardContainer(
            title: chain.name,
            subtitle: chain.description,
            leading: Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: _getChainModeColor(chain.mode).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                _getChainModeIcon(chain.mode),
                color: _getChainModeColor(chain.mode),
              ),
            ),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  '${chain.agents.length} agents',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(width: 8),
                IconButton(
                  icon: const Icon(Icons.play_arrow),
                  onPressed: () => _executeChain(chain),
                  tooltip: 'Execute chain',
                ),
              ],
            ),
            onTap: () => _showChainDetails(chain),
          );
        },
      ),
    );
  }

  Widget _buildTemplatesTab() {
    if (_templates.isEmpty) {
      return const EmptyState(
        title: 'No Templates',
        message: 'Templates help you quickly create common agent chains',
        icon: Icons.view_module,
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _templates.length,
      itemBuilder: (context, index) {
        final template = _templates[index];
        return CardContainer(
          title: template.name,
          subtitle: template.description,
          leading: const Icon(Icons.view_module, size: 32),
          trailing: ElevatedButton(
            onPressed: () => _instantiateTemplate(template),
            child: const Text('Use Template'),
          ),
        );
      },
    );
  }

  Widget _buildExecutionsTab() {
    if (_activeExecutions.isEmpty) {
      return const EmptyState(
        title: 'No Active Executions',
        message: 'Execute a chain to see it here',
        icon: Icons.pending_actions,
      );
    }

    return RefreshIndicator(
      onRefresh: _loadData,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _activeExecutions.length,
        itemBuilder: (context, index) {
          final execution = _activeExecutions[index];
          return CardContainer(
            title: 'Chain: ${execution.chainId}',
            subtitle: 'Started ${_formatDateTime(execution.startTime)}',
            leading: _buildExecutionStatusIcon(execution.status),
            trailing: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  execution.status.toUpperCase(),
                  style: TextStyle(
                    color: _getStatusColor(execution.status),
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (execution.errors.isNotEmpty)
                  Text(
                    '${execution.errors.length} errors',
                    style: TextStyle(
                      color: Colors.red,
                      fontSize: 12,
                    ),
                  ),
              ],
            ),
            onTap: () => _showExecutionDetails(execution),
          );
        },
      ),
    );
  }

  void _showCreateChainDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Create Agent Chain'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Choose how to create your chain:'),
            const SizedBox(height: 16),
            ListTile(
              leading: const Icon(Icons.build),
              title: const Text('Build from scratch'),
              subtitle: const Text('Create a custom chain'),
              onTap: () {
                Navigator.of(context).pop();
                _showBuildChainDialog();
              },
            ),
            ListTile(
              leading: const Icon(Icons.view_module),
              title: const Text('Use template'),
              subtitle: const Text('Start with a predefined template'),
              onTap: () {
                Navigator.of(context).pop();
                setState(() => _selectedTab = 1); // Switch to templates tab
              },
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );
  }

  void _showBuildChainDialog() {
    // In a real app, this would open a chain builder UI
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Chain builder coming soon!')),
    );
  }

  void _executeChain(AgentChain chain) async {
    // Show input dialog for chain parameters
    final TextEditingController controller = TextEditingController();
    
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Execute ${chain.name}'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('Provide initial data for the chain:'),
            const SizedBox(height: 16),
            TextField(
              controller: controller,
              decoration: const InputDecoration(
                labelText: 'Input (e.g., search query, goal)',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.of(context).pop({
                'input': controller.text,
              });
            },
            child: const Text('Execute'),
          ),
        ],
      ),
    );

    if (result != null && result['input']?.isNotEmpty == true) {
      final request = ChainExecutionRequest(
        chainId: chain.id!,
        initialData: result,
      );

      try {
        final execution = await _service.executeChain(request);
        if (execution != null) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Chain execution started: ${execution.id}'),
              action: SnackBarAction(
                label: 'View',
                onPressed: () {
                  setState(() => _selectedTab = 2);
                  _loadData();
                },
              ),
            ),
          );
        }
      } catch (e) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to execute chain: $e')),
        );
      }
    }
  }

  void _showChainDetails(AgentChain chain) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(chain.name),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(chain.description),
              const SizedBox(height: 16),
              Text(
                'Execution Mode: ${chain.mode}',
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
              Text('Timeout: ${chain.timeoutSeconds}s'),
              const SizedBox(height: 16),
              const Text(
                'Agents:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              ...chain.agents.map((agent) => Padding(
                padding: const EdgeInsets.only(left: 16, top: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('• ${agent.name} (${agent.type})'),
                    Text(
                      agent.description,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              )),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Close'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.of(context).pop();
              _executeChain(chain);
            },
            child: const Text('Execute'),
          ),
        ],
      ),
    );
  }

  void _showExecutionDetails(ChainExecutionResult execution) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Execution ${execution.id.substring(0, 8)}...'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Chain: ${execution.chainId}'),
              Text('Status: ${execution.status}'),
              Text('Started: ${_formatDateTime(execution.startTime)}'),
              if (execution.endTime != null)
                Text('Ended: ${_formatDateTime(execution.endTime!)}'),
              const SizedBox(height: 16),
              if (execution.errors.isNotEmpty) ...[
                const Text(
                  'Errors:',
                  style: TextStyle(fontWeight: FontWeight.bold, color: Colors.red),
                ),
                ...execution.errors.map((error) => Padding(
                  padding: const EdgeInsets.only(left: 16, top: 4),
                  child: Text(
                    '${error.agentName}: ${error.error}',
                    style: const TextStyle(color: Colors.red),
                  ),
                )),
                const SizedBox(height: 16),
              ],
              const Text(
                'Execution Log:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              ...execution.executionLog.map((log) => Padding(
                padding: const EdgeInsets.only(left: 16, top: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${log.agentName} (${log.durationSeconds.toStringAsFixed(2)}s)',
                      style: const TextStyle(fontWeight: FontWeight.w500),
                    ),
                    Text(
                      'Status: ${log.status}',
                      style: TextStyle(
                        color: log.status == 'completed' ? Colors.green : Colors.red,
                      ),
                    ),
                  ],
                ),
              )),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  void _instantiateTemplate(ChainTemplate template) async {
    // Show parameter input dialog
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Creating chain from ${template.name} template...')),
    );
    
    // In a real app, this would show a form to fill template parameters
    // For now, we'll use dummy data
    final parameters = {
      'name': '${template.name} - ${DateTime.now().millisecondsSinceEpoch}',
      ...Map.fromEntries(
        template.parameters.map((p) => MapEntry(p.name, 'Sample value')),
      ),
    };

    try {
      final chain = await _service.instantiateTemplate(template.id, parameters);
      if (chain != null) {
        await _loadData();
        setState(() => _selectedTab = 0); // Switch to chains tab
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Chain created successfully!')),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to create chain: $e')),
      );
    }
  }

  Color _getChainModeColor(String mode) {
    switch (mode) {
      case 'sequential':
        return Colors.blue;
      case 'parallel':
        return Colors.green;
      case 'conditional':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  IconData _getChainModeIcon(String mode) {
    switch (mode) {
      case 'sequential':
        return Icons.arrow_forward;
      case 'parallel':
        return Icons.call_split;
      case 'conditional':
        return Icons.alt_route;
      default:
        return Icons.account_tree;
    }
  }

  Widget _buildExecutionStatusIcon(String status) {
    IconData icon;
    Color color;

    switch (status) {
      case 'running':
        icon = Icons.play_circle;
        color = Colors.blue;
        break;
      case 'completed':
        icon = Icons.check_circle;
        color = Colors.green;
        break;
      case 'failed':
        icon = Icons.error;
        color = Colors.red;
        break;
      case 'timeout':
        icon = Icons.timer_off;
        color = Colors.orange;
        break;
      default:
        icon = Icons.help;
        color = Colors.grey;
    }

    return Icon(icon, color: color, size: 32);
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'running':
        return Colors.blue;
      case 'completed':
        return Colors.green;
      case 'failed':
        return Colors.red;
      case 'timeout':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  String _formatDateTime(DateTime dateTime) {
    final now = DateTime.now();
    final difference = now.difference(dateTime);

    if (difference.inMinutes < 1) {
      return 'just now';
    } else if (difference.inHours < 1) {
      return '${difference.inMinutes}m ago';
    } else if (difference.inDays < 1) {
      return '${difference.inHours}h ago';
    } else {
      return '${dateTime.day}/${dateTime.month} ${dateTime.hour}:${dateTime.minute.toString().padLeft(2, '0')}';
    }
  }
}