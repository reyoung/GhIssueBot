# GhIssueBot

Very simple Github Issue Robot. It could watch a github repo, and send email when issue **open** or **reopen** and new issue 
comment **created** .

It use YAML as configuration, the sample configuration is

```YAML
duty:
  mon:
    - WHO_SHOULD_BE_NOTIFIED_ON_MONDAY@gmail.com
    - WHO_ELSE@someother.com
  tue:
    - WHO_SHOULD_BE_NOTIFIED_ON_TUESDAY@gmail.com
  wed:
    - ...
  thurs:
    - ...
  fri:
    - ...
  sat:
    - ...
  sun:
    - ...
email:
  addr: SENDER_EMAIL@gmail.com
  password: SENDER_PASSWD
secretCode: github_webhook_secret_code
```

The email will be send to the users that are on duty. Duty table is devided by weekdays. 

Only GMail is supported by sender.
