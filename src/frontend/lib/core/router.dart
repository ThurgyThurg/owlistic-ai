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
import 'package:owlistic/screens/calendar_screen.dart';
import 'package:owlistic/utils/logger.dart';

class AppRouter {
  final Logger _logger = Logger('AppRouter');
  
  late final GoRouter router;

  AppRouter(BuildContext context) {
    router = GoRouter(
      debugLogDiagnostics: true,
      initialLocation: '/',
      errorBuilder: (context, state) => Scaffold(
        appBar: AppBar(title: const Text('Navigation Error')),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.error_outline,
                size: 64,
                color: Theme.of(context).colorScheme.error,
              ),
              const SizedBox(height: 16),
              Text(
                'Navigation Error',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                'Unable to load the requested page',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => context.go('/'),
                child: const Text('Go to Home'),
              ),
            ],
          ),
        ),
      ),
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
            final String? notebookId = state.pathParameters['id'];
            if (notebookId == null) {
              return Scaffold(
                appBar: AppBar(title: const Text('Error')),
                body: const Center(
                  child: Text('Invalid notebook ID'),
                ),
              );
            }
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
            final String? noteId = state.pathParameters['id'];
            if (noteId == null) {
              return Scaffold(
                appBar: AppBar(title: const Text('Error')),
                body: const Center(
                  child: Text('Invalid note ID'),
                ),
              );
            }
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
        GoRoute(
          path: '/calendar',
          builder: (context, state) => const CalendarScreen(),
        ),
      ],
    );
  }
}

