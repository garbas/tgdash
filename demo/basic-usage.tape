# basic-usage.tape - Show tgdash processing terragrunt output
# Demonstrates: piping terragrunt output, dashboard view,
# units progressing through states

Output site/static/demo/basic-usage.gif

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Build tgdash and set clean prompt
Hide
Type `export PS1="$ "`
Enter
Type "go build -o /tmp/tgdash ."
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms
Show

# Show the command we're about to run
Type "cat internal/integration/testdata/basic_plan.txt | /tmp/tgdash"
Sleep 1s
Enter

# Let the dashboard render and process
Sleep 5s

# Screenshot the final state
Sleep 2s
