# Deltagrams

In this conversation, we will use **deltagrams** for exchanging collections of
file changes and operations from a hierarchical directory structure. Below
are instructions on how to understand and process deltagrams.

## Format Specification

### Structure
- A deltagram consists of one or more parts separated by boundary markers.
- Each part contains headers followed by delta operations.
- Parts are separated by boundary lines starting with `--`.
- The final boundary must repeat the same ==== before the two hyphens: ====--. So the full final line is always --====DELTAGRAM_{uuid}====--.

### Boundary Markers
- Format: `--====DELTAGRAM_{uuid}====`
    - `uuid` is a 32-character hexadecimal UUID, conforming to IETF RFC 4122.
    - Non-hexadecimal or malformed UUIDs (e.g., containing letters beyond f or
      incorrect length) are invalid and may cause parsing errors.
- Examples:
    - Initial/Partitional: `--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====`
    - Final: `--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====--`

### Headers
Each part must include:
1. `Content-Location`:
   - For optional messages: `deltagram://message`
   - For file operations: original filesystem path or URL
2. `Content-Type`: Operation type and metadata
   - For content deltas: `application/x-deltagram-content; charset=utf-8; linesep=LF`
   - For file operations: `application/x-deltagram-fileop; charset=utf-8`
3. `Delta-Operation`: The type of delta operation
   - `content`: Content modifications (add/delete/replace lines)
   - `create`: Create new file
   - `delete`: Delete file
   - `move`: Move/rename file
   - `copy`: Copy file

Omitting any required header renders the part invalid.

### Delta Operations

#### Operation Selection Guidelines

**CRITICAL**: Always choose the correct operation type:

- **Use `create`** for new files that don't exist
- **Use `content`** ONLY for modifying existing files  
- **Never use `content` on non-existent files** - this will cause runtime errors

**File Existence Check**: Before using `content` operation, ensure the target file exists in the current project state. If uncertain, use `create` operation instead.

**Content Accuracy Check**: Verify that the unified diff matches the actual current content of the target file. Line numbers and content must correspond exactly to avoid application errors.

#### Content Modifications
For `Delta-Operation: content`, the body contains line-based delta operations:

```
@@ -start_line,count +start_line,count @@
-deleted line content
+added line content
 unchanged line content (context)
```

Operations:
- `-` prefix: Delete this line
- `+` prefix: Add this line
- ` ` prefix: Context line (unchanged)
- `@@ -old_start,old_count +new_start,new_count @@`: Hunk header

#### File Operations
For file operations, the body contains operation-specific data:

**Create File** (`Delta-Operation: create`):
```
+++ /path/to/new/file.txt
file content here
```

**Delete File** (`Delta-Operation: delete`):
```
--- /path/to/file.txt
```

**Move/Rename File** (`Delta-Operation: move`):
```
--- /old/path/file.txt
+++ /new/path/file.txt
```

**Copy File** (`Delta-Operation: copy`):
```
--- /source/path/file.txt
+++ /destination/path/file.txt
```

### Content
- Follows headers after a blank line.
- Normalized to UTF-8 character set.
- Normalized to Unix (LF) line endings.
- Delta operations use unified diff format for content changes.

## Interpretation Guidelines

### Messages
- Parts with `Content-Location: deltagram://message` contain human messages.
- These messages provide context about the changes or may be a general
  response to a previous assistant turn.

### File Parts
- Represent changes to text files from a filesystem or URL.
- Content-Location paths may be:
  - Relative paths (e.g., `src/main.py`)
  - Absolute paths (e.g., `/home/user/project/main.py`)
  - URLs (e.g., `https://example.com/file.txt`)
- Paths maintain their hierarchy even in the flat bundle format.

### Delta Application Order
1. File deletions are applied first
2. File moves/renames are applied second
3. File copies are applied third  
4. File creations are applied fourth
5. Content modifications are applied last

## Example

```
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Added logging functionality and fixed bug in main function.
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====
Content-Location: src/logger.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ src/logger.py
import logging
import sys

class Logger:
    def __init__(self, name="app"):
        self.logger = logging.getLogger(name)
        self.logger.setLevel(logging.INFO)
        
        handler = logging.StreamHandler(sys.stdout)
        formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
        handler.setFormatter(formatter)
        self.logger.addHandler(handler)
    
    def info(self, message):
        self.logger.info(message)
    
    def error(self, message):
        self.logger.error(message)
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====
Content-Location: src/main.py
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -1,5 +1,8 @@
+from logger import Logger
+
 def main():
+    logger = Logger()
+    logger.info("Starting application")
     print("Hello, world!")
-    return 0
+    return True
 
 if __name__ == "__main__":
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====--
```

In this example:
1. The message part explains the purpose of the changes.
2. The first delta creates a new logger.py file.
3. The second delta modifies main.py to add logging and fix the return type.

## Processing Instructions

When working with deltagrams:
1. Read the message first, if there is one, to understand context.
2. Apply deltas in the specified order (deletions, moves, copies, creations, content changes).
3. Examine file paths to understand project structure changes.
4. Maintain the same format when responding with file changes.
5. Preserve original paths (in `Content-Location`) unless explicitly asked to change them.
6. Use unified diff format for content modifications.

*Important*:
- Only provide responses as deltagrams when the user explicitly requests them as such. Otherwise, provide responses as you normally would.
- If you have an artifacts or canvases mechanism available, then write deltagrams via this mechanism. Otherwise, display them as normal responses.
- When writing a deltagram to an artifact or canvas, treat it as plain text rather than Markdown.
- Always validate that delta operations can be applied successfully before generating the deltagram.

## ✅ Final Output Validation for Deltagrams

Before delivering any deltagram, **always verify these six points**:

1. **Boundary Markers**
   - Each part must start with: `--====DELTAGRAM_{uuid}====`
   - The **final boundary** must end with `====--` (two trailing hyphens).
     Example: `--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====--`

2. **UUID Format**
   - Every `{uuid}` must be exactly 32 lowercase hexadecimal characters (0–9, a–f).
   - No non-hex letters, no uppercase, no extra or missing characters.

3. **Required Headers**
   - Each part must contain:
     - `Content-Location:`
     - `Content-Type:`
     - `Delta-Operation:` (for non-message parts)
   - All must appear before the blank line and content.

4. **Line Endings**
   - All lines must match the specified `linesep` (e.g., `linesep=LF` for Unix line endings).
   - No stray `\r` or mixed line endings.

5. **No Placeholders**
   - Do not output placeholder UUIDs like `aaaaaaaa...` or `bbbbbbbb...`.
   - Always generate valid, unique UUIDs for production deltagrams.

**6. Valid Delta Operations**
   - Content deltas must use proper unified diff format.
   - **Line Number Validation**: Verify that all hunk headers (`@@ -start,count +start,count @@`) reference valid line numbers within the target file's actual length
   - **Content Matching**: Ensure that context lines (prefixed with ` `) exactly match the current file content at the specified line numbers
   - File operation syntax must be correct (+++ for additions, --- for deletions/sources).
   - Hunk headers must accurately reflect line numbers and counts.
   - All referenced files must exist or be created within the same deltagram.

## Efficiency Benefits

Deltagrams provide several advantages over full file transmission:

1. **Reduced Size**: Only changes are transmitted, not entire files.
2. **Clear Intent**: Each change is explicitly documented with its purpose.
3. **Atomic Operations**: Related changes are grouped together logically.
4. **Conflict Detection**: Easier to identify and resolve conflicting changes.
5. **Version Control Friendly**: Delta format aligns with git diff output.
6. **Selective Application**: Individual changes can be applied or rejected independently.

**Always validate that delta operations can be applied successfully before generating the deltagram.**