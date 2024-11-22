### Features of the Runner Application Ordered by Logical Sequence of Actions

1. **Configuration Management**:
   - Reads the predefined set of services and their configurations from a JSON file.
   - Future feature: Ability to receive configurations from an endpoint.

2. **Dynamic Service Compilation**:
   - Services are compiled (`go build`) before execution to ensure up-to-date binaries.

3. **Service Metadata**:
   - Each service is defined with attributes such as `Path`, `Port`, `Description`, and `BinaryName`.
   - Enables detailed service categorization, logging, and debugging.

4. **Start/Stop Services**:
   - Starts a predefined set of web services immediately upon execution.
   - Supports starting additional services and taking specific services down during runtime.
   - Stops all running services upon receiving a stop command.

5. **Build-Time Configurations**:
   - Allows service-specific environment variables to be set during the build process for better customization.

6. **Logging**:
   - Logs events such as services starting, stopping, and encountering errors.
   - Logging level (e.g., info, warnings, errors, debug) configurable as a startup parameter.

7. **Command Interface**:
   - Accepts commands (e.g., start, stop, add, remove) via HTTP endpoints.

8. **Extensible Endpoint System**:
   - Supports both text and JSON responses for API endpoints, allowing flexible additions and custom APIs.
   - Includes an endpoint (`/eps`) that dynamically lists all available commands and their descriptions.

9. **Real-Time Feedback**:
   - Provides a dashboard accessible via an endpoint to display the status of services.
   - Includes an endpoint (`/eps`) that lists all HTTP endpoints and their metadata.

10. **Graceful Shutdown**:
    - Implements a timeout mechanism for shutting down services cleanly.
    - Uses process group IDs for sending signals to entire groups of processes, ensuring clean termination.

11. **Error Handling**:
    - Logs and reports errors during service start or stop operations.
    - Future feature: Real-time error reporting via endpoints or dashboard.

12. **Default Port Management**:
    - Server uses the `SERVER_PORT` environment variable to set the port, with a fallback to `8080`.

13. **Service Cleanup**:
    - Automatically cleans up binaries after service execution to minimize clutter in directories.

14. **Custom Module Path Handling**:
    - Configures `GOMODCACHE` and `GOPATH` paths dynamically, ensuring compatibility with custom environments.

15. **Scalability**:
    - Designed to manage approximately 10 services initially.
    - Scalable to handle up to 1000 services in the future.

16. **Health Monitoring**:
    - Future feature: Monitors service health and restarts crashed services.

17. **Security**:
    - Open access to endpoints for now.
    - Includes a stub for authentication and authorization to allow future expansion.
