import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../viewmodel/theme_viewmodel.dart';
import '../viewmodel/settings_viewmodel.dart';
import '../widgets/app_bar_common.dart';
import '../widgets/card_container.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({Key? key}) : super(key: key);

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
}