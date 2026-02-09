# error-filtering.tape - Show error detection and filtering
# Demonstrates: errors highlighted, e key to filter errors

Output site/static/demo/error-filtering.gif

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Build and set clean prompt
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

# Use the fixture with errors
Type "cat internal/integration/testdata/plan_with_errors.txt | /tmp/tgdash"
Sleep 1s
Enter

# Let dashboard render with errors visible
Sleep 3s

# Navigate to see the error
Type "j"
Sleep 500ms

# Toggle error filter with e
Type "e"
Sleep 2s

# Show only errors are visible now
Sleep 2s

# Toggle back to see all units
Type "e"
Sleep 2s
