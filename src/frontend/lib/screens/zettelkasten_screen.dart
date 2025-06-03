import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:owlistic/providers/zettelkasten_provider.dart';
import 'package:owlistic/widgets/app_bar_common.dart';
import 'package:owlistic/widgets/zettelkasten_graph_widget.dart';
import 'package:owlistic/widgets/zettelkasten_side_panel.dart';
import 'package:owlistic/models/zettelkasten.dart';
import 'package:owlistic/utils/logger.dart';

/// Main screen for the Zettelkasten knowledge graph visualization
class ZettelkastenScreen extends StatefulWidget {
  const ZettelkastenScreen({super.key});

  @override
  State<ZettelkastenScreen> createState() => _ZettelkastenScreenState();
}

class _ZettelkastenScreenState extends State<ZettelkastenScreen> {
  final Logger _logger = Logger('ZettelkastenScreen');
  
  bool _showSidePanel = true;
  ZettelNode? _selectedNode;
  String _selectedFilter = 'all';
  List<String> _selectedTags = [];
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadGraphData();
    });
  }

  Future<void> _loadGraphData() async {
    final provider = context.read<ZettelkastenProvider>();
    await provider.loadGraphData();
  }

  void _onNodeSelected(ZettelNode? node) {
    setState(() {
      _selectedNode = node;
    });
  }

  void _onFilterChanged(String filter) {
    setState(() {
      _selectedFilter = filter;
    });
    _applyFilters();
  }

  void _onTagsChanged(List<String> tags) {
    setState(() {
      _selectedTags = tags;
    });
    _applyFilters();
  }

  void _onSearchChanged(String query) {
    setState(() {
      _searchQuery = query;
    });
    _applyFilters();
  }

  void _applyFilters() {
    final provider = context.read<ZettelkastenProvider>();
    
    ZettelSearchInput? filter;
    if (_selectedFilter != 'all' || _selectedTags.isNotEmpty || _searchQuery.isNotEmpty) {
      filter = ZettelSearchInput(
        query: _searchQuery.isEmpty ? null : _searchQuery,
        nodeTypes: _selectedFilter == 'all' ? null : [_selectedFilter],
        tags: _selectedTags.isEmpty ? null : _selectedTags,
      );
    }
    
    provider.applyFilter(filter);
  }

  void _toggleSidePanel() {
    setState(() {
      _showSidePanel = !_showSidePanel;
    });
  }

  Future<void> _syncAllContent() async {
    final provider = context.read<ZettelkastenProvider>();
    final result = await provider.syncAllContent();
    
    if (result != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(result['message'] ?? 'Content synchronized successfully'),
          backgroundColor: Colors.green,
        ),
      );
      await provider.loadGraphData();
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Failed to synchronize content'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  Future<void> _analyzeGraph() async {
    final provider = context.read<ZettelkastenProvider>();
    final analysis = await provider.analyzeGraph();
    
    if (analysis != null && mounted) {
      _showAnalysisDialog(analysis);
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Failed to analyze graph'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _showAnalysisDialog(Map<String, dynamic> analysis) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(analysis['title'] ?? 'Graph Analysis'),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (analysis['description'] != null) ...[
                Text(
                  analysis['description'],
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 16),
              ],
              if (analysis['insights'] != null) ...[
                Text(
                  'Insights:',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                const SizedBox(height: 8),
                for (final insight in analysis['insights'])
                  Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Text('• $insight'),
                  ),
                const SizedBox(height: 16),
              ],
              if (analysis['recommendations'] != null) ...[
                Text(
                  'Recommendations:',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                const SizedBox(height: 8),
                for (final recommendation in analysis['recommendations'])
                  Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Text('• $recommendation'),
                  ),
              ],
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBarCommon(
        title: 'Knowledge Graph',
        actions: [
          if (!_showSidePanel)
            IconButton(
              icon: const Icon(Icons.filter_list),
              onPressed: _toggleSidePanel,
              tooltip: 'Show filters',
            ),
          IconButton(
            icon: const Icon(Icons.sync),
            onPressed: _syncAllContent,
            tooltip: 'Sync all content',
          ),
          IconButton(
            icon: const Icon(Icons.analytics),
            onPressed: _analyzeGraph,
            tooltip: 'Analyze graph',
          ),
          PopupMenuButton<String>(
            onSelected: (value) {
              switch (value) {
                case 'export':
                  _exportGraph();
                  break;
                case 'help':
                  _showHelpDialog();
                  break;
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'export',
                child: ListTile(
                  leading: Icon(Icons.download),
                  title: Text('Export Graph'),
                  contentPadding: EdgeInsets.zero,
                ),
              ),
              const PopupMenuItem(
                value: 'help',
                child: ListTile(
                  leading: Icon(Icons.help),
                  title: Text('Help'),
                  contentPadding: EdgeInsets.zero,
                ),
              ),
            ],
          ),
        ],
      ),
      body: Consumer<ZettelkastenProvider>(
        builder: (context, provider, child) {
          if (provider.isLoading) {
            return const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  CircularProgressIndicator(),
                  SizedBox(height: 16),
                  Text('Loading knowledge graph...'),
                ],
              ),
            );
          }

          if (provider.errorMessage != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.error,
                    size: 64,
                    color: Theme.of(context).colorScheme.error,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Error loading graph',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    provider.errorMessage!,
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _loadGraphData,
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (provider.graph == null || provider.graph!.nodes.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.account_tree,
                    size: 64,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'No graph data',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Sync your content to create the knowledge graph',
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _syncAllContent,
                    child: const Text('Sync Content'),
                  ),
                ],
              ),
            );
          }

          return Row(
            children: [
              // Main graph area
              Expanded(
                child: ZettelkastenGraphWidget(
                  graph: provider.filteredGraph ?? provider.graph!,
                  selectedNode: _selectedNode,
                  onNodeSelected: _onNodeSelected,
                  onNodePositionChanged: (nodeId, position) {
                    provider.updateNodePosition(nodeId, position);
                  },
                  onConnectionCreated: (sourceId, targetId, connectionType) {
                    provider.createConnection(sourceId, targetId, connectionType);
                  },
                  onConnectionDeleted: (edgeId) {
                    provider.deleteConnection(edgeId);
                  },
                ),
              ),
              
              // Side panel
              if (_showSidePanel)
                Container(
                  width: 300,
                  decoration: BoxDecoration(
                    border: Border(
                      left: BorderSide(
                        color: Theme.of(context).dividerColor,
                      ),
                    ),
                  ),
                  child: ZettelkastenSidePanel(
                    selectedNode: _selectedNode,
                    availableTags: provider.graph?.tags ?? [],
                    selectedFilter: _selectedFilter,
                    selectedTags: _selectedTags,
                    onFilterChanged: _onFilterChanged,
                    onTagsChanged: _onTagsChanged,
                    onNodeSelected: _onNodeSelected,
                    onSearchChanged: _onSearchChanged,
                    onClose: _toggleSidePanel,
                  ),
                ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _exportGraph() async {
    final provider = context.read<ZettelkastenProvider>();
    final graph = await provider.exportGraph();
    
    if (graph != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Graph exported: ${graph.nodes.length} nodes, ${graph.edges.length} connections',
          ),
          backgroundColor: Colors.green,
        ),
      );
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Failed to export graph'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _showHelpDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Knowledge Graph Help'),
        content: const SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Navigation:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              Text('• Pan: Click and drag on empty space'),
              Text('• Zoom: Mouse wheel or pinch gesture'),
              Text('• Select node: Click on a node'),
              SizedBox(height: 16),
              Text(
                'Interactions:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              Text('• Move nodes: Drag selected nodes'),
              Text('• Create connections: Use the side panel'),
              Text('• Filter: Use the side panel filters'),
              SizedBox(height: 16),
              Text(
                'Node Types:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              Text('• Blue: Notes'),
              Text('• Green: Tasks'),
              Text('• Orange: Projects'),
              SizedBox(height: 16),
              Text(
                'Connection Types:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              Text('• Related: General relationship'),
              Text('• Depends On: Dependency relationship'),
              Text('• References: Citation or reference'),
              Text('• Supports: Supporting evidence'),
              Text('• Contradicts: Conflicting information'),
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
}