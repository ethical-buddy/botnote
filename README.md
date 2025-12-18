# 🗒️ MyNotes CLI

A lightweight, terminal-based productivity tool for Hyprland users. Manage your tasks and write long-form notes using your system's native `$EDITOR` (Vim/Neovim).

## ✨ Features
- **Categorized Todos:** Clear separation between Pending and Completed tasks.
- **Deep Writing:** Notes are written in Vim, allowing for large content and markdown support.
- **Side-by-Side UI:** View your tasks and note titles in a single split-pane view.
- **Background Daemon:** Sends desktop notifications via `notify-send` when tasks are due.

## ⌨️ Controls
| Key | Action |
| :--- | :--- |
| `Tab` | Switch between Todo and Note panes |
| `t` | Create a new Task |
| `n` | Create a new Note (Title first, then Vim) |
| `v` | View/Edit selected Note in Vim |
| `x` | Delete selected Task or Note |
| `Enter` | Toggle Task completion |
| `q` | Quit |

## 🚀 Installation
1. Ensure you have Go and `notify-send` (libnotify) installed.
2. Run the install script:
   ```bash
   chmod +x install.sh
   ./install.sh
