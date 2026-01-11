# List subcommand

Display executables managed by pathman. 
 - Optionally filters by name (exact match).
 - Optionally add complete details rather than just a summary. 
 - Optionally output in json, full details always given.

Usage:
  pathman list [executable-name] [flags]

Aliases:
  list, ls

Flags:
  -h, --help              help for list
      --json              Output as JSON
  -l, --long              Show detailed information
      --priority string   List only from 'front' or 'back' folder
      --type string       List only file or directory entries
      --bypriority        Order by priority rather than by type (file/directory)

Global Flags:
      --version   Print version information

## Output Format Specification

The general concept is that the format is reminiscent of the /usr/bin/ls command. This format specification should be replicated across execman, scriptman, and pathman for consistency.

### Compact Format (default)

The compact format lists only the executable names, one per line, with no headers or additional information.

Example output:
```
decisions
my-script
another-tool
```

**Characteristics:**
- No headers or footers
- No extra information (repo, version, etc.)
- Just executable names, one per line
- Suitable for piping to other commands
- Sorted 
    - Partitioned by file/directory type and alphabetically sorted within that (default)
    - Or by priority order, --bypriority ignored by --json

### Long Format (--long, -l)

The long format shows all available information in a human-friendly labeled format. Each executable is separated by a blank line.

Example output:
```
File:         execman
Symlink:      /home/sfkleach/.local/libexec/execman
Priority:     front

Directory:    /home/sfkleach/go/bin/
Priority:     back
```

**Characteristics:**
- Left-aligned labels in a fixed width with colon separator
 for readability
- All available metadata displayed
- Blank line between entries

### JSON Format (--json)

The JSON format outputs the data structure for programmatic consumption. The
data is partitioned by type and within that sorted alphabetically.

Example output:
```json
{
    "files": [
        {
            "file": "execman",
            "symlink": "/home/sfkleach/.local/libexec/execman",
            "priority": "front"
        }
    ],
    "directories": [
        {
            "directory": "/home/sfkleach/go/bin/",
            "priority": "back"
        }
    ]
}
```

**Characteristics:**
- Standard JSON formatting
- Complete data structure
- Suitable for parsing by other tools

### Filtering Behavior

When an executable name is provided as an argument, the output is filtered to show only exact matches i.e. not partial matches such as a prefix. The format remains the same (compact, long, or JSON) but only includes matching executables. 
