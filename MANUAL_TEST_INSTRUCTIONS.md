# Manual Integration Test for Corrupted Storage Recovery

This document provides instructions for manually testing the corrupted storage recovery feature (task 4.3).

## Prerequisites

1. Build the binary: `just build`
2. The test file `test_corrupted_entries.jsonl` contains a mix of valid and corrupted entries

## Test Procedure

### Test 1: Verify Binary Builds Successfully

```bash
just build
```

**Expected:** Binary builds without errors and creates `./did` executable

### Test 2: Test with Corrupted Storage File

1. **Backup your existing storage** (if you have one):
   ```bash
   # On Linux
   cp ~/.config/did/entries.jsonl ~/.config/did/entries.jsonl.backup
   ```

2. **Copy the test file to storage location**:
   ```bash
   # On Linux
   mkdir -p ~/.config/did
   cp test_corrupted_entries.jsonl ~/.config/did/entries.jsonl
   ```

3. **Run the `did` command** (list today's entries):
   ```bash
   ./did
   ```

   **Expected Output:**
   - **stderr** should show warnings about corrupted lines:
     ```
     Warning: Found 3 corrupted line(s) in storage file:
       Line 3: {"Timestamp":"2026-01-02T12:00:00Z","Descr... (error: ...)
       Line 5: {invalid json syntax here} (error: ...)
       Line 7: truncated line without proper JSON {"Timesta... (error: ...)
     ```
   - **stdout** should show valid entries for today (entries 1-5)
   - Total duration should be calculated correctly from valid entries only

4. **Run the `did validate` command**:
   ```bash
   ./did validate
   ```

   **Expected Output:**
   - Storage file path displayed
   - Total lines: 9
   - Valid entries: 6
   - Corrupted entries: 3
   - Detailed list of corrupted lines with line numbers and content
   - Status: ⚠ Storage file has 3 corrupted line(s)

### Test 3: Test with Healthy Storage File

1. **Create a clean test file**:
   ```bash
   cat > ~/.config/did/entries.jsonl <<EOF
   {"Timestamp":"2026-01-02T10:00:00Z","Description":"Clean entry 1","DurationMinutes":30,"RawInput":"Clean entry 1 for 30m"}
   {"Timestamp":"2026-01-02T11:00:00Z","Description":"Clean entry 2","DurationMinutes":60,"RawInput":"Clean entry 2 for 1h"}
   EOF
   ```

2. **Run validate command**:
   ```bash
   ./did validate
   ```

   **Expected Output:**
   - Total lines: 2
   - Valid entries: 2
   - Corrupted entries: 0
   - Status: ✓ Storage file is healthy

### Test 4: Test Other Commands with Corrupted Storage

Using the corrupted test file from Test 2:

```bash
./did y    # Yesterday's entries
./did w    # This week's entries
./did lw   # Last week's entries
```

**Expected:** All commands should show warnings about corrupted lines on stderr, but still display valid entries filtered by their respective time ranges.

### Test 5: Test Adding New Entry with Corrupted Storage

```bash
./did "test entry" for 15m
```

**Expected:**
- New entry should be appended successfully
- Success message should be displayed
- Running `./did validate` should show 1 more valid entry

## Cleanup

Restore your original storage file:
```bash
# On Linux
mv ~/.config/did/entries.jsonl.backup ~/.config/did/entries.jsonl
```

Or remove the test data:
```bash
rm ~/.config/did/entries.jsonl
```

## Acceptance Criteria Verification

After completing the tests, verify all acceptance criteria are met:

- [x] **Malformed JSON lines are skipped with a warning to stderr** - Verified in Test 2
- [x] **Valid entries before and after corrupted lines are still loaded** - Verified in Test 2
- [x] **Warning includes the line number and content of corrupted entries** - Verified in Test 2
- [x] **did validate command shows storage file health status** - Verified in Test 2 and Test 3
- [x] **Corrupted entries don't cause crashes or data loss for valid entries** - Verified across all tests

## Notes

- Warnings are always written to **stderr** (standard error)
- Valid entries and success messages are written to **stdout** (standard output)
- The tool continues to function normally even with corrupted storage
- Line numbers in warnings are 1-indexed (first line is line 1)
- Content in warnings is truncated to 50 characters for readability
