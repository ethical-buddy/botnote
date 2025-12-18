#!/bin/bash

# 1. Build the application
echo "Building bot the todo notes app..."
go build -o bot main.go

# 2. Move to local bin
echo "Installing to ~/.local/bin..."
mkdir -p ~/.local/bin
mv bot ~/.local/bin/

# 3. Create Systemd Service for Daemon
echo "Setting up background daemon..."
mkdir -p ~/.config/systemd/user/

cat <<EOF >~/.config/systemd/user/bot-daemon.service
[Unit]
Description=bot Notification Daemon
After=graphical-session.target

[Service]
ExecStart=%h/.local/bin/bot daemon
Restart=always

[Install]
WantedBy=graphical-session.target
EOF

# 4. Reload and Enable
systemctl --user daemon-reload
systemctl --user enable bot-daemon.service
systemctl --user restart bot-daemon.service

echo "Done! You can now run the tool by typing 'mynotes' in your terminal."
echo "The notification daemon is running in the background."
