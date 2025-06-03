import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:owlistic/core/theme.dart';
import 'package:owlistic/viewmodel/home_viewmodel.dart';
import 'app_logo.dart';

class SidebarDrawer extends StatelessWidget {
  const SidebarDrawer({Key? key}) : super(key: key);

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
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/');
            },
          ),
          ListTile(
            leading: const Icon(Icons.book),
            title: const Text('Notebooks'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/notebooks');
            },
          ),
          ListTile(
            leading: const Icon(Icons.description),
            title: const Text('Notes'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/notes');
            },
          ),
          ListTile(
            leading: const Icon(Icons.task),
            title: const Text('Tasks'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/tasks');
            },
          ),
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('AI Dashboard'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/ai-dashboard');
            },
          ),
          ListTile(
            leading: const Icon(Icons.account_tree),
            title: const Text('Knowledge Graph'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/zettelkasten');
            },
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.delete),
            title: const Text('Trash'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/trash');
            },
          ),
          ListTile(
            leading: const Icon(Icons.settings),
            title: const Text('Settings'),
            onTap: () {
              Navigator.pop(context);
              GoRouter.of(context).go('/settings');
            },
          ),
          const Spacer(),
        ],
      ),
    );
  }

}
