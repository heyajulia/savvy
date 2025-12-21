# Agent Guidelines

## Commit Messages

- Make sure commit messages are clear and descriptive. Always consult `git log` to see previous commit messages for
  style reference.
- Follow the 50-72 rule: subject line under 50 characters, body wrapped at 72.
- End the subject line with a period.
- Do not include promotional lines like "🤖 Generated with [Claude Code]" or similar.
- AI agents should identify themselves using a `Co-authored-by` trailer identifying the AI model and version plus your
  GitHub email address, e.g.:

  ```
  Co-authored-by: Claude Opus 4.5 <noreply@anthropic.com>
  ```
