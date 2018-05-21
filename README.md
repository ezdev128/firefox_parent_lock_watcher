# firefox_parent_lock_watcher
Watches over .parentlock file being created by Firefox and removes it.

### Why?
Mozilla Firefox creates the .parentlock file in the current profile directory, e.g. (~/.mozilla/firefox/profilename/.parentlock) and prevents opening new instance of Firefox.

### Is it safe?
Yes.

### Compiling and installing
go build

go install

### Lanuch
/path/to/firefox_parent_lock_watcher /path/to/firefox/profile.ini
