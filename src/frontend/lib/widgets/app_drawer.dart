import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:owlistic/core/theme.dart';
import 'package:owlistic/viewmodel/home_viewmodel.dart';
import 'app_logo.dart';

class SidebarDrawer extends StatelessWidget {
  const SidebarDrawer({Key? key}) : super(key: key);

  void _navigateTo(BuildContext context, String route) {
    // Close drawer first, then navigate after a brief delay to avoid conflicts
    Navigator.pop(context);
    Future.microtask(() {
      if (context.mounted) {
        context.go(route);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 16),
            decoration: BoxDecoration(
              color: Theme.of(context).primaryColor,
            ),
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                AppLogo(size: 32, forceTransparent: true),
                SizedBox(width: 12),
                Text(
                  'Owlistic',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
          ListTile(
            leading: const Icon(Icons.home),
            title: const Text('Home'),
            onTap: () => _navigateTo(context, '/'),
          ),
          ListTile(
            leading: const Icon(Icons.book),
            title: const Text('Notebooks'),
            onTap: () => _navigateTo(context, '/notebooks'),
          ),
          ListTile(
            leading: const Icon(Icons.description),
            title: const Text('Notes'),
            onTap: () => _navigateTo(context, '/notes'),
          ),
          ListTile(
            leading: const Icon(Icons.task),
            title: const Text('Tasks'),
            onTap: () => _navigateTo(context, '/tasks'),
          ),
          ListTile(
            leading: const Icon(Icons.calendar_today),
            title: const Text('Calendar'),
            onTap: () => _navigateTo(context, '/calendar'),
          ),
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('AI Dashboard'),
            onTap: () => _navigateTo(context, '/ai-dashboard'),
          ),
          ListTile(
            leading: const Icon(Icons.account_tree),
            title: const Text('Knowledge Graph'),
            onTap: () => _navigateTo(context, '/zettelkasten'),
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.delete),
            title: const Text('Trash'),
            onTap: () => _navigateTo(context, '/trash'),
          ),
          ListTile(
            leading: const Icon(Icons.settings),
            title: const Text('Settings'),
            onTap: () => _navigateTo(context, '/settings'),
          ),
          const Spacer(),
        ],
      ),
    );
  }

}
