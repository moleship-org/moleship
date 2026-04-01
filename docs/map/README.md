# Project Path

## List of Initial Goals

1. Reading and Status Interface (Observability)

    * List of Quadlets: Read the ~/.config/containers/systemd/ directory and display all detected .container, .network, and .volume files.
    * Status mapping: Check the actual status of each service associated with the Quadlet using `systemctl --user is-active <service>` and display it visually.
    * Inspection: Connect to the Podman socket to cross-check data and display whether the underlying container is actually running and which ports are assigned to it.

2. Lifecycle Control (Basic Operation)

    * Status Actions: Implement quick-action buttons for each Quadlet: Start, Stop, and Restart (which execute the corresponding systemctl --user commands in the background).
    * Daemon Reload: A global “Sync” or “Reload” button that executes `systemctl --user daemon-reload` so the system recognizes if you’ve manually modified a `.container` file from the console.

3. Debugging (Troubleshooting)
    * Log Viewer: A dashboard that retrieves the latest 50–100 lines of the log for a specific service using `journalctl --user -u <service>`. This is essential for determining why a duvet-stack container or proxy failed to start properly.

4. Architecture and Basic Security
    * Standalone backend: Build the API, ensuring that system calls are sanitized to prevent code injection.
