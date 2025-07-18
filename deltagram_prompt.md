# Deltagrams

## Quick-Start Checklist for LLMs

**Follow these steps every time you generate a deltagram:**

1. **Select the correct operation for each file:**
   - **New file or uncertain if exists** → Use `create`
   - **File definitely exists and you know its exact content** → Use `content`
   - **When in doubt** → Use `create`

2. **For `content` operations with multiple hunks:**
   - **Use original file line numbers** for all hunks
   - **System automatically handles offsets** - no manual calculation needed
   - Reference the file as it exists before any changes

3. **Generate a valid boundary identifier:**
   - Must be at least 8 characters using alphanumeric, underscore, or dash (a-z, A-Z, 0-9, _, -)
   - Can be a UUID (32 hex chars) or other format like `voice456_sample-789012345678901234ef`
   - Example: `083f1e1306624ef4a246c23193d3fdd7` or `voice456_sample-789012345678901234ef`

4. **Size limit:**
   - Each deltagram must be ≤ 4,000 characters
   - Split into multiple batches if needed

5. **Always include a message part:**
   - `Content-Location: deltagram://message`
   - Summarize intent and batch info if applicable

6. **File parts must be plain text:**
   - No Markdown formatting in file content

7. **Validate before output:**
   - Check boundaries, identifiers, headers, and operations

8. All deltagrams must be generated as artifacts. Never output deltagram content directly in chat.

## Common Mistakes to Avoid

1. **Using `content` on non-existent files**
   - Error: "cannot apply content operation to non-existent file"
   - Solution: Use `create` instead

2. **Markdown in file parts**
   - File content must be plain text, not Markdown

3. **Invalid boundary identifiers**
   - Must be at least 8 characters using alphanumeric, underscore, or dash (a-z, A-Z, 0-9, _, -)
   - Can be UUIDs (32 hex chars) or other formats for flexibility

4. **Exceeding size limits**
   - Split into multiple batches if over 4,000 characters

5. **Missing required headers**
   - All parts need `Content-Location` and `Content-Type`
   - Non-message parts need `Delta-Operation`

6. **Wrong final boundary**
   - Must end with `====--` (two trailing hyphens)

7. Using `content` without verifying file state**
   - Error: "diff attempts to remove line X but file only has Y lines"
   - Solution: Check document contents or use `create` instead
   - **Rule: Use `create` when changing >50% of file content**

---

## Format Specification

### Structure Overview
- Parts separated by boundary markers: `--====DELTAGRAM_{uuid}====`
- Final boundary ends with `====--`
- Each part has headers followed by content
- Operations applied in specific order

### Boundary Markers
- **Format:** `--====DELTAGRAM_{identifier}====`
- **Identifier:** At least 8 characters using alphanumeric, underscore, or dash (a-z, A-Z, 0-9, _, -)
- **Final boundary:** `--====DELTAGRAM_{identifier}====--`

**Example:**
```
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====
--====DELTAGRAM_voice456_sample-789012345678901234ef====
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====--
```

### Required Headers

**All parts must include:**
- `Content-Location`: Path or `deltagram://message`
- `Content-Type`: Operation type and metadata

**Non-message parts must also include:**
- `Delta-Operation`: The operation type

**Header Examples:**
```
Content-Location: src/main.py
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content
```

## Operation Selection Guide

### When to Use Each Operation

**Use `create` when:**
- Creating a new file
- Uncertain if file exists
- **Replacing >50% of file content**
- **When in doubt, always use `create`**

**Use `content` when:**
- File definitely exists
- Making small, specific edits (<50% of content)
- **You've verified exact current content from documents**

### Before Using `content` Operations
- **Check document contents**: Verify exact line counts and content
- **For major changes**: Use `create` instead of complex diffs
- **When uncertain**: Always prefer `create` over `content`

**Use `delete` when:**
- Removing an existing file

**Use `move` when:**
- Renaming or moving a file

**Use `copy` when:**
- Duplicating a file to a new location

### Operation Formats

#### Create File (`create`)
```
Content-Location: path/to/file.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ path/to/file.txt
file content here
```

#### Modify Content (`content`)
```
Content-Location: path/to/file.txt
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -2,3 +2,4 @@
 unchanged line
-removed line
+added line
 unchanged line
```

#### Delete File (`delete`)
```
Content-Location: path/to/file.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: delete

--- path/to/file.txt
```

#### Move/Rename File (`move`)
```
Content-Location: old/path/file.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: move

--- old/path/file.txt
+++ new/path/file.txt
```

#### Copy File (`copy`)
```
Content-Location: source/path/file.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: copy

--- source/path/file.txt
+++ destination/path/file.txt
```

## Unified Diff Format (for `content` operations)

### Automatic Offset Calculation
**IMPORTANT**: The deltagram system automatically handles line number offsets for multiple hunks. You can reference line numbers from the original file, and the system will automatically calculate where subsequent hunks should be applied after previous changes.

**Example**: If the first hunk adds 3 lines, and your second hunk references line 20 in the original file, the system will automatically apply it at the correct position (line 23 in the modified file).

### Hunk Header
```
@@ -old_start,old_count +new_start,new_count @@
```

### Line Prefixes
- `+` Add this line
- `-` Remove this line
- ` ` (space) Context line (unchanged)

### Example
```
@@ -1,5 +1,6 @@
 import os
+import sys
 
 def main():
-    print("Hello")
+    print("Hello, world!")
     return 0
```

### Multiple Hunks with Automatic Offsets
When creating multiple hunks for the same file, reference the original file line numbers:
```
@@ -5,2 +5,3 @@
 def setup():
+    # Add initialization
     return True

@@ -10,1 +11,2 @@
 def main():
+    setup()
     print("Hello")
```
The system automatically adjusts the second hunk's position based on changes from the first hunk.

## Batching for Large Changes

### When to Split
- Individual deltagram exceeds 4,000 characters
- Logical grouping of related changes
- Performance considerations

### Batching Rules
- Each batch must be a complete, valid deltagram
- Include batch info in message part
- Apply batches in sequence order
- Same file can appear in multiple batches

### Message Format for Batches
```
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Batch 1 of 3: Core module updates
Implementing authentication system changes.
Apply batches in order.
```

## Application Order

Operations are applied in this sequence:
1. **Delete** files
2. **Move/rename** files
3. **Copy** files
4. **Create** new files
5. **Modify content** of existing files

## Complete Example

```
--====DELTAGRAM_083f1e1306624ef4a246c23193d3fdd7====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Adding logging functionality and fixing main function return value.
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
        formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
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

## Final Validation Checklist

**Before outputting any deltagram, verify:**

1. **Boundary markers are correct:**
   - Start: `--====DELTAGRAM_{identifier}====`
   - End: `--====DELTAGRAM_{identifier}====--`

2. **Identifier is valid:**
   - At least 8 characters using alphanumeric, underscore, or dash (a-z, A-Z, 0-9, _, -)

3. **All required headers present:**
   - `Content-Location`
   - `Content-Type`
   - `Delta-Operation` (for non-message parts)

4. **Content format is correct:**
   - Unified diff format for `content` operations
   - Proper file operation syntax
   - No Markdown in file parts

5. **Size under limit:**
   - Total deltagram ≤ 4,000 characters

6. **Operations are valid:**
   - Line numbers reference the original file (automatic offset calculation handles subsequent hunks)
   - Referenced files exist or are created in same deltagram

## Output Guidelines

- **Only provide deltagrams when explicitly requested**
- **Use artifacts/canvases if available**
- **Treat deltagrams as plain text, not Markdown**
- **Always validate before delivery**
- **If uncertain about file existence, use `create` and explain in message**

## Benefits of Deltagrams

1. **Reduced size** - Only changes transmitted
2. **Clear intent** - Each change documented
3. **Atomic operations** - Related changes grouped
4. **Conflict detection** - Easier to identify issues
5. **Version control friendly** - Aligns with git diff
6. **Selective application** - Individual changes can be applied/rejected