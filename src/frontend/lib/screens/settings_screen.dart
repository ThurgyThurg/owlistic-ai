import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../viewmodel/theme_viewmodel.dart';
import '../viewmodel/settings_viewmodel.dart';
import '../widgets/app_bar_common.dart';
import '../widgets/card_container.dart';
import '../providers/calendar_provider.dart';
import '../services/calendar_service.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
                    trailing: Consumer<ThemeViewModel>(
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
      builder: (context) => ChangeNotifierProvider(
        create: (_) => SettingsViewModel(),
        child: Consumer<SettingsViewModel>(
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
          final isConnected = ref.watch(googleCalendarConnectedProvider);
          final calendarService = ref.watch(calendarServiceProvider);
          
          return AlertDialog(
            title: const Row(
              children: [
                Icon(Icons.calendar_today),
                SizedBox(width: 8),
                Text('Google Calendar'),
              ],
            ),
            content: Column(
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
                        ref.read(googleCalendarConnectedProvider.notifier).state = false;
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
                              // Show dialog to enter auth code
                              _showAuthCodeDialog(context, ref);
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

  void _showAuthCodeDialog(BuildContext context, WidgetRef ref) {
    final authCodeController = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Enter Authorization Code'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              'After authorizing in your browser, copy the authorization code and paste it here:',
              style: TextStyle(fontSize: 14),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: authCodeController,
              decoration: const InputDecoration(
                labelText: 'Authorization Code',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () async {
              if (authCodeController.text.isNotEmpty) {
                try {
                  final calendarService = ref.read(calendarServiceProvider);
                  await calendarService.connectGoogleCalendar(authCodeController.text);
                  ref.read(googleCalendarConnectedProvider.notifier).state = true;
                  if (context.mounted) {
                    Navigator.of(context).pop();
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Successfully connected to Google Calendar')),
                    );
                  }
                } catch (e) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Error: $e')),
                    );
                  }
                }
              }
            },
            child: const Text('Connect'),
          ),
        ],
      ),
    );
  }
}