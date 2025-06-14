import 'package:flutter/material.dart';
import 'package:provider/provider.dart' as Provider;
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../viewmodel/theme_viewmodel.dart';
import '../viewmodel/settings_viewmodel.dart';
import '../widgets/app_bar_common.dart';
import '../widgets/card_container.dart';
import '../providers/calendar_provider.dart';
import '../services/calendar_service.dart';

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  @override
  void initState() {
    super.initState();
    
    // Check for calendar connection success from OAuth redirect
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final uri = Uri.parse(ModalRoute.of(context)?.settings.name ?? '');
      if (uri.queryParameters.containsKey('calendar_connected')) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Google Calendar connected successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        // Refresh the connection status
        ref.invalidate(googleCalendarConnectedProvider);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const AppBarCommon(title: 'Settings'),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            CardContainer(
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.palette),
                    title: const Text('Appearance'),
                    subtitle: const Text('Customize app theme'),
                    trailing: Provider.Consumer<ThemeViewModel>(
                      builder: (context, themeViewModel, child) {
                        return Switch(
                          value: themeViewModel.isDarkMode,
                          onChanged: (value) {
                            themeViewModel.toggleTheme();
                          },
                        );
                      },
                    ),
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.auto_awesome),
                    title: const Text('AI Features'),
                    subtitle: const Text('AI-powered note enhancement'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _showAIFeaturesDialog(context);
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.notifications),
                    title: const Text('Notifications'),
                    subtitle: const Text('Manage notification preferences'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _showNotificationSettingsDialog(context);
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.calendar_today),
                    title: const Text('Google Calendar'),
                    subtitle: const Text('Connect your Google Calendar'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _showGoogleCalendarDialog(context);
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            CardContainer(
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.person),
                    title: const Text('Account'),
                    subtitle: const Text('Manage your profile and security'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      context.go('/profile');
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.storage),
                    title: const Text('Data & Storage'),
                    subtitle: const Text('Manage storage and export data'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _showDataStorageDialog(context);
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            CardContainer(
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.info),
                    title: const Text('About'),
                    subtitle: const Text('Version 1.0.0'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _showAboutDialog(context);
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.help),
                    title: const Text('Help & Support'),
                    subtitle: const Text('Get help and report issues'),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      // TODO: Implement help screen
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showAIFeaturesDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.auto_awesome),
            SizedBox(width: 8),
            Text('AI Features'),
          ],
        ),
        content: const Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Available AI features:'),
            SizedBox(height: 8),
            Text('• Automatic title generation'),
            Text('• Content summarization'),
            Text('• Smart tag extraction'),
            Text('• Related note discovery'),
            SizedBox(height: 16),
            Text(
              'Look for the ✨ button in note editors to enhance your notes with AI.',
              style: TextStyle(fontStyle: FontStyle.italic),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Got it'),
          ),
        ],
      ),
    );
  }

  void _showAboutDialog(BuildContext context) {
    showAboutDialog(
      context: context,
      applicationName: 'Owlistic',
      applicationVersion: '1.0.0',
      applicationIcon: const Icon(Icons.auto_stories, size: 48),
      children: const [
        Text('A powerful note-taking app with AI-powered features.'),
      ],
    );
  }

  void _showNotificationSettingsDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => Provider.ChangeNotifierProvider(
        create: (_) => SettingsViewModel(),
        child: Provider.Consumer<SettingsViewModel>(
          builder: (context, viewModel, child) {
            return AlertDialog(
              title: const Row(
                children: [
                  Icon(Icons.notifications),
                  SizedBox(width: 8),
                  Text('Notification Settings'),
                ],
              ),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (viewModel.isLoading)
                      const Center(child: CircularProgressIndicator())
                    else ...[
                      if (viewModel.errorMessage != null)
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.red.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            viewModel.errorMessage!,
                            style: const TextStyle(color: Colors.red),
                          ),
                        ),
                      SwitchListTile(
                        title: const Text('Enable Notifications'),
                        subtitle: const Text('Receive all app notifications'),
                        value: viewModel.notificationsEnabled,
                        onChanged: (value) {
                          viewModel.toggleNotifications(value);
                        },
                      ),
                      const Divider(),
                      SwitchListTile(
                        title: const Text('Note Reminders'),
                        subtitle: const Text('Get reminded about important notes'),
                        value: viewModel.noteReminders,
                        onChanged: viewModel.notificationsEnabled
                            ? (value) {
                                viewModel.toggleNoteReminders(value);
                              }
                            : null,
                      ),
                      SwitchListTile(
                        title: const Text('Task Reminders'),
                        subtitle: const Text('Notifications for upcoming tasks'),
                        value: viewModel.taskReminders,
                        onChanged: viewModel.notificationsEnabled
                            ? (value) {
                                viewModel.toggleTaskReminders(value);
                              }
                            : null,
                      ),
                      SwitchListTile(
                        title: const Text('AI Insights'),
                        subtitle: const Text('Smart suggestions and insights'),
                        value: viewModel.aiInsights,
                        onChanged: viewModel.notificationsEnabled
                            ? (value) {
                                viewModel.toggleAIInsights(value);
                              }
                            : null,
                      ),
                      const SizedBox(height: 16),
                      const Text(
                        'Note: Notification settings are synced across all your devices.',
                        style: TextStyle(fontSize: 12, fontStyle: FontStyle.italic),
                      ),
                    ],
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Done'),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  void _showDataStorageDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.storage),
            SizedBox(width: 8),
            Text('Data & Storage'),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Storage Options:'),
            const SizedBox(height: 16),
            ListTile(
              leading: const Icon(Icons.download),
              title: const Text('Export All Data'),
              subtitle: const Text('Download your notes as JSON'),
              onTap: () {
                Navigator.of(context).pop();
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Export feature coming soon')),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.backup),
              title: const Text('Backup Settings'),
              subtitle: const Text('Configure automatic backups'),
              onTap: () {
                Navigator.of(context).pop();
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Backup settings coming soon')),
                );
              },
            ),
            const SizedBox(height: 16),
            const Text(
              'Note: Your data is automatically synced across all your devices.',
              style: TextStyle(fontSize: 12, fontStyle: FontStyle.italic),
            ),
          ],
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

  void _showGoogleCalendarDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => Consumer(
        builder: (context, ref, child) {
          final connectionAsync = ref.watch(googleCalendarConnectedProvider);
          final calendarService = ref.watch(calendarServiceProvider);
          
          return AlertDialog(
            title: const Row(
              children: [
                Icon(Icons.calendar_today),
                SizedBox(width: 8),
                Text('Google Calendar'),
              ],
            ),
            content: connectionAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (error, stack) => Text('Error checking connection: $error'),
              data: (isConnected) => Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (isConnected) ...[
                    const Row(
                      children: [
                        Icon(Icons.check_circle, color: Colors.green, size: 20),
                        SizedBox(width: 8),
                        Text('Connected to Google Calendar'),
                      ],
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'Your calendar events are synced with Google Calendar.',
                      style: TextStyle(fontSize: 14),
                    ),
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton.icon(
                        onPressed: () async {
                          // Show calendar selection dialog
                          _showCalendarSelectionDialog(context, ref, calendarService);
                        },
                        icon: const Icon(Icons.settings),
                        label: const Text('Manage Calendar Sync'),
                      ),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed: () async {
                          await calendarService.syncWithGoogle();
                          if (context.mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Calendar synced successfully')),
                            );
                          }
                        },
                        icon: const Icon(Icons.sync),
                        label: const Text('Sync Now'),
                      ),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton.icon(
                        onPressed: () async {
                          await calendarService.disconnectGoogleCalendar();
                          ref.invalidate(googleCalendarConnectedProvider);
                          if (context.mounted) {
                            Navigator.of(context).pop();
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Disconnected from Google Calendar')),
                            );
                          }
                        },
                        icon: const Icon(Icons.link_off),
                        label: const Text('Disconnect'),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.red,
                        ),
                      ),
                    ),
                  ] else ...[
                  const Text(
                    'Connect your Google Calendar to:',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  const Text('• Sync your events across devices'),
                  const Text('• Create events in Google Calendar'),
                  const Text('• Get reminders and notifications'),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.blue.withOpacity(0.3)),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(Icons.info, color: Colors.blue, size: 20),
                            const SizedBox(width: 8),
                            Text(
                              'Google Cloud Console Setup Required',
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.blue[800],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        const Text(
                          'Before connecting, add this callback URL to your Google Cloud Console:',
                          style: TextStyle(fontSize: 14),
                        ),
                        const SizedBox(height: 8),
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.grey[100],
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: FutureBuilder<Map<String, dynamic>>(
                                  future: calendarService.getOAuthConfig(),
                                  builder: (context, snapshot) {
                                    if (snapshot.hasData) {
                                      return Text(
                                        snapshot.data!['redirect_uri'] ?? 'Loading...',
                                        style: TextStyle(
                                          fontFamily: 'monospace',
                                          fontSize: 12,
                                          color: Colors.grey[800],
                                        ),
                                      );
                                    } else {
                                      return Text(
                                        'Loading configuration...',
                                        style: TextStyle(
                                          fontFamily: 'monospace',
                                          fontSize: 12,
                                          color: Colors.grey[600],
                                        ),
                                      );
                                    }
                                  },
                                ),
                              ),
                              IconButton(
                                icon: const Icon(Icons.copy, size: 16),
                                onPressed: () {
                                  // TODO: Implement clipboard copy
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(
                                      content: Text('Copy the URL manually for now'),
                                      duration: Duration(seconds: 2),
                                    ),
                                  );
                                },
                                tooltip: 'Copy URL',
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          '1. Go to Google Cloud Console → APIs & Services → Credentials\n'
                          '2. Edit your OAuth 2.0 Client ID\n'
                          '3. Add the above URL to "Authorized redirect URIs"\n'
                          '4. Save and try connecting again',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[700],
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      onPressed: () async {
                        try {
                          final authUrl = await calendarService.getGoogleAuthUrl();
                          if (await canLaunchUrl(Uri.parse(authUrl))) {
                            await launchUrl(Uri.parse(authUrl));
                            if (context.mounted) {
                              Navigator.of(context).pop();
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Please complete the authorization in your browser. Return here and refresh to check connection status.'),
                                  duration: Duration(seconds: 5),
                                ),
                              );
                            }
                          }
                        } catch (e) {
                          if (context.mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(content: Text('Error: $e')),
                            );
                          }
                        }
                      },
                      icon: const Icon(Icons.link),
                      label: const Text('Connect Google Calendar'),
                    ),
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
          );
        },
      ),
    );
  }

  void _showCalendarSelectionDialog(BuildContext context, WidgetRef ref, CalendarService calendarService) {
    showDialog(
      context: context,
      builder: (context) => FutureBuilder<List<Map<String, dynamic>>>(
        future: calendarService.listGoogleCalendars(),
        builder: (context, snapshot) {
          return AlertDialog(
            title: const Row(
              children: [
                Icon(Icons.calendar_month),
                SizedBox(width: 8),
                Text('Select Calendars to Sync'),
              ],
            ),
            content: SizedBox(
              width: double.maxFinite,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (snapshot.connectionState == ConnectionState.waiting)
                    const Center(child: CircularProgressIndicator())
                  else if (snapshot.hasError)
                    Text(
                      'Error loading calendars: ${snapshot.error}',
                      style: const TextStyle(color: Colors.red),
                    )
                  else if (!snapshot.hasData || snapshot.data!.isEmpty)
                    const Text('No calendars found')
                  else
                    Expanded(
                      child: FutureBuilder<Map<String, dynamic>>(
                        future: calendarService.getSyncStatus(),
                        builder: (context, syncSnapshot) {
                          final syncedCalendars = <String>{};
                          if (syncSnapshot.hasData) {
                            final syncList = syncSnapshot.data!['sync_calendars'] as List? ?? [];
                            for (var sync in syncList) {
                              syncedCalendars.add(sync['google_calendar_id']);
                            }
                          }
                          
                          return ListView.builder(
                            shrinkWrap: true,
                            itemCount: snapshot.data!.length,
                            itemBuilder: (context, index) {
                              final calendar = snapshot.data![index];
                              final calendarId = calendar['id'] as String;
                              final calendarName = calendar['summary'] as String? ?? 'Unnamed Calendar';
                              final isPrimary = calendar['primary'] == true;
                              final isSynced = syncedCalendars.contains(calendarId);
                              
                              return Card(
                                child: ListTile(
                                  leading: Icon(
                                    Icons.calendar_today,
                                    color: isSynced ? Colors.green : null,
                                  ),
                                  title: Text(calendarName),
                                  subtitle: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      if (isPrimary)
                                        const Text(
                                          'Primary Calendar',
                                          style: TextStyle(fontWeight: FontWeight.bold),
                                        ),
                                      if (isSynced)
                                        const Text(
                                          'Currently synced',
                                          style: TextStyle(color: Colors.green),
                                        ),
                                    ],
                                  ),
                                  trailing: isSynced
                                      ? const Icon(Icons.check_circle, color: Colors.green)
                                      : ElevatedButton(
                                          onPressed: () async {
                                            try {
                                              await calendarService.setupCalendarSync(
                                                calendarId: calendarId,
                                                calendarName: calendarName,
                                              );
                                              if (context.mounted) {
                                                Navigator.of(context).pop();
                                                ScaffoldMessenger.of(context).showSnackBar(
                                                  SnackBar(
                                                    content: Text('Set up sync for $calendarName'),
                                                  ),
                                                );
                                                // Trigger initial sync
                                                await calendarService.syncWithGoogle();
                                              }
                                            } catch (e) {
                                              if (context.mounted) {
                                                ScaffoldMessenger.of(context).showSnackBar(
                                                  SnackBar(
                                                    content: Text('Error: $e'),
                                                    backgroundColor: Colors.red,
                                                  ),
                                                );
                                              }
                                            }
                                          },
                                          child: const Text('Sync This'),
                                        ),
                                ),
                              );
                            },
                          );
                        },
                      ),
                    ),
                  const SizedBox(height: 16),
                  const Text(
                    'Select which Google Calendars to sync with Owlistic.',
                    style: TextStyle(fontSize: 12, fontStyle: FontStyle.italic),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: const Text('Done'),
              ),
            ],
          );
        },
      ),
    );
  }

}