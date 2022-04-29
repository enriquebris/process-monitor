# process-monitor

Monitor all active processes and kill those who match the "*processes*" entry (config.json)

### configuration file: config.json

```json
{
  "processes": [
    {
      "name_regex": "^Spotify$",
      "cron": "@every 5s",
      "cpu_max_limit": 2,
      "kill_if_cpu_max_limit": true,
      "total_attempts_before_kill": 3
    },
    {
      "name_regex": "^Activity Monitor$",
      "cron": "@every 5s",
      "cpu_max_limit": 2,
      "kill_if_cpu_max_limit": true,
      "total_attempts_before_kill": 3
    }
  ]
}
```

#### cron expression format

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

#### predefined schedules

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *

#### intervals

```
@every <duration>
```
