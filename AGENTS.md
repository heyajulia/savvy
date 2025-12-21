# Agent Guidelines

## Commit Messages

- Make sure commit messages are clear and descriptive. Always consult a command like
  `git log -15 --format='%H%n%s%n%n%b%n===' --author-date-order` to see previous commit messages for style reference.
  Don't just look at the subject line; read the full commit message to get a sense of tone and level of detail.
- Follow the 50-72 rule: subject line under 50 characters, body wrapped at 72.
- End the subject line with a period.
- Do not include promotional lines like "🤖 Generated with [Claude Code]" or similar.
- AI agents should identify themselves using a `Co-authored-by` trailer identifying the AI model and version plus your
  GitHub email address, e.g.:

  ```
  Co-authored-by: Claude Opus 4.5 <noreply@anthropic.com>
  ```
