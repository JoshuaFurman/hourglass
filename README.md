# hourglass

`hourglass` is a cross-platform (linux/MacOS) cli/tui timer that utilizes native desktop notifications.

# Using hourglass on MacOS

`hourglass` uses the https://github.com/gen2brain/beeep/ library to send desktop notifications. This library uses `osascript` on MacOS under the hood.
if notifications are not being displayed properly you may need to allow `osascript` notifications. To do so, follow these steps:

- open `script editor` on your Mac.
- initiate a dummy notification. Go to File->New and paste the follwing `display notification "hello" with title "hi"` and click the Play icon ( run the script )
- you will get a notification to accept/allow notifications from script editor
- Solution reference: https://forum.latenightsw.com/t/trying-to-use-terminal-for-display-notification/5068
- beep library issue reference: https://github.com/gen2brain/beeep/issues/67

