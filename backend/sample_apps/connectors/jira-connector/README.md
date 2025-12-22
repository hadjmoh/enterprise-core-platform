# Jira Connector

Sync security incidents with Jira tickets and track SLA compliance.

## Features

- Automatic ticket creation for incidents
- Bi-directional sync (incident ↔ ticket)
- SLA tracking and compliance monitoring
- Workflow status updates
- Comment synchronization

## Configuration

- **Jira URL**: Your Jira instance URL
- **API Token**: Jira API token
- **Project Key**: Target project (e.g., SEC)
- **Issue Type**: Default issue type (e.g., Bug, Task)
- **Auto-Create**: Automatically create tickets for new incidents

## Permissions Required

- `read:incidents` - View security incidents
- `write:data` - Write Jira ticket data to indexes

## Dashboard

Monitor ticket management with:
- Ticket status distribution
- SLA compliance rate
- Incident to ticket mapping

## Installation

1. Generate Jira API token
2. Install connector from App Store
3. Configure Jira credentials and project
4. Enable auto-create if desired
5. View dashboard

## Support

Contact Enterprise Security Team for assistance.
