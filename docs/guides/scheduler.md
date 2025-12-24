# Scheduler Guide

Manage scheduled jobs for formations.

## Overview

The scheduler allows you to create recurring jobs that send messages to your formation on a schedule. Jobs use cron syntax for flexible scheduling.

## View Scheduler Config

View the scheduler configuration:

```bash
muxi scheduler
```

Output:
```
Scheduler Configuration

Enabled:     true
Max Jobs:    100
Timezone:    UTC
```

## List Jobs

List all scheduled jobs:

```bash
muxi scheduler jobs
muxi scheduler jobs -u alice    # Filter by user
```

Output:
```
Scheduled Jobs (3)

ID                    TYPE      SCHEDULE        NEXT RUN              USER
job_abc123            message   0 9 * * *       Dec 25, 2024 09:00   alice
job_def456            message   */30 * * * *    Dec 24, 2024 14:30   bob
job_ghi789            message   0 0 * * 1       Dec 30, 2024 00:00   alice
```

## Create Job

Create a new scheduled job:

```bash
muxi scheduler create -u alice --schedule "0 9 * * *" --message "Good morning! What's on the agenda today?"
```

### Cron Syntax

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sun-Sat)
│ │ │ │ │
* * * * *
```

### Common Schedules

```bash
# Every day at 9 AM
--schedule "0 9 * * *"

# Every hour
--schedule "0 * * * *"

# Every 30 minutes
--schedule "*/30 * * * *"

# Every Monday at 10 AM
--schedule "0 10 * * 1"

# First day of every month at midnight
--schedule "0 0 1 * *"

# Weekdays at 8:30 AM
--schedule "30 8 * * 1-5"
```

## View Job Details

View details of a specific job:

```bash
muxi scheduler job job_abc123
```

Output:
```
Job Details

ID:          job_abc123
Type:        message
Schedule:    0 9 * * *
User:        alice
Message:     Good morning! What's on the agenda today?
Created:     Dec 20, 2024
Next Run:    Dec 25, 2024 09:00 UTC
Last Run:    Dec 24, 2024 09:00 UTC
Run Count:   4
```

## Delete Job

Delete a scheduled job:

```bash
muxi scheduler delete job_abc123
```

## Options

```
-f, --formation    Formation ID
-p, --profile      Server profile
-u, --user         User ID (required for create)
    --schedule     Cron schedule expression
    --message      Message to send
```

## Use Cases

### Daily Standup Reminder

```bash
muxi scheduler create -u team \
  --schedule "0 9 * * 1-5" \
  --message "Time for standup! What did you work on yesterday?"
```

### Weekly Report

```bash
muxi scheduler create -u alice \
  --schedule "0 17 * * 5" \
  --message "Generate my weekly activity report"
```

### Hourly Check-in

```bash
muxi scheduler create -u monitor \
  --schedule "0 * * * *" \
  --message "Run system health check"
```

## Best Practices

1. **Use descriptive messages** - Make it clear what the job should trigger
2. **Consider timezones** - Jobs run in the formation's configured timezone
3. **Start with longer intervals** - Test with less frequent schedules first
4. **Monitor job history** - Check run counts and last run times
5. **Clean up unused jobs** - Delete jobs that are no longer needed
