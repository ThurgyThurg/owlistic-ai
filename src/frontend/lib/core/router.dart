import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import 'package:owlistic/screens/home_screen.dart';
import 'package:owlistic/screens/notebooks_screen.dart';
import 'package:owlistic/screens/notebook_detail_screen.dart';
import 'package:owlistic/screens/notes_screen.dart';
import 'package:owlistic/screens/note_editor_screen.dart';
import 'package:owlistic/screens/tasks_screen.dart';
import 'package:owlistic/screens/trash_screen.dart';
import 'package:owlistic/screens/user_profile_screen.dart';
import 'package:owlistic/screens/settings_screen.dart';
import 'package:owlistic/screens/ai_dashboard_screen.dart';
import 'package:owlistic/screens/zettelkasten_screen.dart';
import 'package:owlistic/utils/logger.dart';

class AppRouter {
  final Logger _logger = Logger('AppRouter');
  
  late final GoRouter router;

  AppRouter(BuildContext context) {
    router = GoRouter(
      debugLogDiagnostics: true,
      initialLocation: '/',
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => const HomeScreen(),
        ),
        // Login route removed - single user system with external auth
        GoRoute(
          path: '/notebooks',
          builder: (context, state) {
            print('🔥 ROUTER: Building /notebooks route'); // Force console output
            try {
              return const NotebooksScreen();
            } catch (e) {
              print('🔥 ROUTER ERROR: $e'); // Force console output
              // Return a simple error screen instead of crashing
              return Scaffold(
                appBar: AppBar(title: const Text('Error')),
                body: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text('Error loading notebooks screen:'),
                      Text(e.toString()),
                      ElevatedButton(
                        onPressed: () => Navigator.of(context).pop(),
                        child: const Text('Go Back'),
                      ),
                    ],
                  ),
                ),
              );
            }
          },
        ),
        GoRoute(
          path: '/notebooks/:id',
          builder: (context, state) {
            final String notebookId = state.pathParameters['id']!;
            return NotebookDetailScreen(notebookId: notebookId);
          },
        ),
        GoRoute(
          path: '/notes',
          builder: (context, state) => const NotesScreen(),
        ),
        GoRoute(
          path: '/notes/:id',
          builder: (context, state) {
            final String noteId = state.pathParameters['id']!;
            return NoteEditorScreen(noteId: noteId);
          },
        ),
        GoRoute(
          path: '/tasks',
          builder: (context, state) => const TasksScreen(),
        ),
        GoRoute(
          path: '/trash',
          builder: (context, state) => const TrashScreen(),
        ),
        GoRoute(
          path: '/profile',
          builder: (context, state) => const UserProfileScreen(),
        ),
        GoRoute(
          path: '/settings',
          builder: (context, state) => const SettingsScreen(),
        ),
        GoRoute(
          path: '/ai-dashboard',
          builder: (context, state) => const AIDashboardScreen(),
        ),
        GoRoute(
          path: '/zettelkasten',
          builder: (context, state) => const ZettelkastenScreen(),
        ),
      ],
    );
  }
}

