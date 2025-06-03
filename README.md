<div align="center">

  <img src="src/frontend/assets/logo/owlistic-w-text.png" width="400px" />
  
  # 🦉 Open-source real-time notetaking & todo app 🔄⚡️🚀

  [![Release](https://img.shields.io/github/release/owlistic-notes/owlistic)](https://github.com/owlistic-notes/owlistic/releases/latest)
  [![Docs](https://img.shields.io/badge/docs-online-blue.svg)](https://owlistic-notes.github.io/owlistic/docs/category/overview)
  [![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)
  [![standard-readme](https://img.shields.io/badge/standard--readme-OK-green.svg)](https://github.com/RichardLitt/standard-readme)

  [![Activity](https://img.shields.io/github/commit-activity/m/owlistic-notes/owlistic)](https://github.com/owlistic-notes/owlistic/pulse)

</div>

## Table of Contents

- [Quick Links](#quick-links)
- [Features](#features)
- [Screenshots](#screenshots)
- [AI Features](#ai-features)
- [Installation](#install)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Quick Links

- [Installation](https://owlistic-notes.github.io/owlistic/docs/category/installation)
- [Quick Start](https://owlistic-notes.github.io/owlistic/docs/overview/quick-start)
- [FAQ](https://owlistic-notes.github.io/owlistic/docs/troubleshooting/faq)
- [Roadmap](https://owlistic-notes.github.io/owlistic/roadmap)
<!--
- [Api Reference](https://owlistic-notes.github.io/owlistic/docs/category/api-reference)
-->

> [!WARNING]
> Owlistic is still under active development, so you may encounter bugs or breaking changes as we improve.

## Features

- 📒 Notebooks/Notes tree
- ✏️ Rich (WYSIWYG) editor
- ✔️ Inline todo items
- 🔄 Real-time sync
- 🔑 JWT-based auth
- 🔒 Role-based access control
- 🗑 Trash
- 🌓 Dark/Light mode
- ⬇️ Import markdown note
- 🤖 AI-powered task breakdown and project planning
- 🎯 Intelligent note enhancement with action steps and learning items
- 📊 AI dashboard for managing projects and agents

Please have a look at the [features](https://owlistic-notes.github.io/owlistic/docs/category/features) for details.

### Screenshots
<details>
<summary>Show</summary>

### General

| Real Time Updates |
|:---|
| <img src='./docs/website/static/img/screenshots/real_time_updates.gif' width="75%" title="Real Time updates" /> |

### Editor

| Editor | Scrolling | Toolbar |
|:---|:---|:---|
| <img src='./docs/website/static/img/screenshots/editor/editor.png' width="50%" title="Editor Screen" /> | <img src='./docs/website/static/img/screenshots/editor/note_scrolling.gif' width="50%" title="Editor Scrolling" /> | <img src='./docs/website/static/img/screenshots/editor/editor_toolbar.png' width="50%" title="Editor Toolbar" /> |

### Screens

| Home | Sidebar | Profile | Trash |
|:---|:---|:---|:---|
| <img src='./docs/website/static/img/screenshots/home.png' width="50%" title="Home Screen" /> | <img src='./docs/website/static/img/screenshots/sidebar.png' width="50%" title="Home Screen" /> | <img src='./docs/website/static/img/screenshots/profile/profile.png' width="50%" title="Profile Screen" /> | <img src='./docs/website/static/img/screenshots/trash/trash.png' width="50%" title="Trash Screen" /> | 

| Notebooks | Notes | Tasks |
|:---|:---|:---|
| <img src='./docs/website/static/img/screenshots/notebooks/notebooks.png' width="50%" title="Notebooks Screen" /> | <img src='./docs/website/static/img/screenshots/notes/notes.png' width="50%" title="Notes Screen" /> | <img src='./docs/website/static/img/screenshots/tasks/tasks.png' width="50%" title="Tasks Screen" /> |

</details>

## AI Features

Owlistic now includes powerful AI capabilities to enhance your note-taking and productivity workflow. All AI features are powered by Anthropic's Claude and include intelligent fallbacks for robust operation.

### 🤖 AI Dashboard

The AI Dashboard serves as the central hub for all AI-powered features:

- **Task Breakdown**: Enter any goal or complex task and AI will break it down into manageable, sequential steps
- **Project Management**: Convert task breakdowns into trackable projects with metadata and progress tracking
- **Agent History**: View and manage all AI agent runs and their execution status
- **Calendar Integration Ready**: Steps include scheduling placeholders for future Google Calendar integration

### 🎯 AI-Enhanced Notes

Transform your notes with intelligent AI insights:

- **Smart Summaries**: AI generates concise summaries of your note content
- **Action Steps**: Automatically extract actionable tasks from your notes
- **Learning Opportunities**: Identify concepts and knowledge gaps for further exploration
- **AI Tags**: Intelligent tagging based on content analysis
- **Related Notes**: Semantic search to find related content across your knowledge base
- **Collapsible Interface**: Clean, organized display of AI insights that can be expanded/collapsed

### 🧠 Agentic AI System

Built on a robust agent architecture:

- **AI Agent Models**: Complete data structures for agent runs, steps, and project management
- **Agent Types**: Support for different agent behaviors (task breakdown, goal planning, reasoning loops)
- **Status Tracking**: Real-time monitoring of agent execution with progress indicators
- **Error Handling**: Robust error recovery and fallback mechanisms
- **Background Processing**: Long-running AI tasks don't block the user interface

### 🔧 Technical Implementation

- **Backend Models**: Complete Go models for AI agents, projects, and enhanced notes
- **REST API**: Full API coverage for all AI operations with proper authentication
- **Vector Embeddings**: ChromaDB integration for semantic search and note similarity
- **Claude Integration**: Anthropic Claude API integration with proper error handling and timeouts
- **Structured Prompting**: Carefully crafted prompts for consistent, high-quality AI responses
- **Fallback Systems**: Graceful degradation when AI services are unavailable

### 🚀 Future Enhancements

- **Google Calendar Integration**: Automatic scheduling of task breakdown steps
- **Agent-to-Agent Communication**: Collaborative AI workflows
- **Learning Adaptation**: AI that learns from your preferences and improves over time
- **Workflow Automation**: Create custom AI workflows for repeated tasks
- **Advanced Analytics**: Insights into your productivity patterns and AI usage

## Install

Spin up Owlistic in minutes using your preferred [installation method](https://owlistic-notes.github.io/owlistic/docs/category/installation).

## Contributing

Owlistic is developed by the community, for the community. We welcome contributions of all kinds - from code improvements to documentation updates. Check out our [Contributing Guide](https://owlistic-notes.github.io/owlistic/docs/category/contributing) to learn how you can help.

Small note: If editing the README, please conform to the
[standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## Contributors

<a href="https://github.com/owlistic-notes/owlistic/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=owlistic-notes/owlistic" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## Star history

[![Star History Chart](https://api.star-history.com/svg?repos=owlistic-notes/owlistic&type=Date)](https://www.star-history.com/#owlistic-notes/owlistic)

## License

GPLv3.0 © 2025 owlistic-notes
