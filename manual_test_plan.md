# Manual Test Plan - Configurable Week Start Day

Test Date: Monday, January 5, 2026

## Acceptance Criteria to Verify

1. ✅ Default remains Monday (ISO week) for backward compatibility
2. ✅ Configuration option: week_start_day = 'sunday' or 'monday'
3. ✅ Week commands respect the configured start day
4. ✅ Headers show the correct date range based on configuration
5. ✅ Works correctly across year boundaries

## Test Scenarios

### Scenario 1: Default Behavior (No Config)
**Expected**: Weeks should start on Monday by default

### Scenario 2: Monday Start Configuration
**Config**: `week_start_day = "monday"`
**Expected**:
- This week: Monday Jan 5 - Sunday Jan 11, 2026
- Last week: Monday Dec 29, 2025 - Sunday Jan 4, 2026 (crosses year boundary!)

### Scenario 3: Sunday Start Configuration
**Config**: `week_start_day = "sunday"`
**Expected**:
- This week: Sunday Jan 5 - Saturday Jan 11, 2026
- Last week: Sunday Dec 28, 2025 - Saturday Jan 4, 2026 (crosses year boundary!)

### Scenario 4: Invalid Configuration
**Config**: `week_start_day = "tuesday"`
**Expected**: Should show validation error

## Test Results

### ✅ Scenario 1: Default Behavior (No Config)
**Test**: Removed config file, tested `did w` and `did lw`
**Result**: PASSED
- Week correctly starts on Monday (Jan 5 - Jan 11, 2026)
- Last week: Dec 29, 2025 - Jan 4, 2026
- Default is Monday (ISO 8601 standard) ✓

### ✅ Scenario 2: Monday Start Configuration (Explicit)
**Test**: Set `week_start_day = "monday"` in config.toml
**Result**: PASSED
- This week: Jan 5 - Jan 11, 2026 ✓
- Last week: Dec 29, 2025 - Jan 4, 2026 ✓
- Year boundary handled correctly ✓

### ✅ Scenario 3: Sunday Start Configuration
**Test**: Set `week_start_day = "sunday"` in config.toml
**Result**: PASSED
- This week: Jan 4 - Jan 10, 2026 ✓
- Last week: Dec 28, 2025 - Jan 3, 2026 ✓
- Year boundary handled correctly ✓
- Stats command also respects Sunday start ✓

### ✅ Scenario 4: Invalid Configuration
**Test**: Set `week_start_day = "tuesday"` in config.toml
**Result**: PASSED
- Clear error message: "invalid week_start_day: must be 'monday' or 'sunday', got 'tuesday'" ✓
- Validation working correctly ✓

## Acceptance Criteria Verification

### ✅ 1. Default remains Monday (ISO week) for backward compatibility
**Status**: PASSED
- Without config file: weeks start on Monday
- With empty config: weeks start on Monday (DefaultConfig returns "monday")
- Existing behavior preserved

### ✅ 2. Configuration option: week_start_day = 'sunday' or 'monday'
**Status**: PASSED
- Both "sunday" and "monday" values work correctly
- Invalid values are rejected with clear error message
- Config is properly loaded and applied

### ✅ 3. Week commands respect the configured start day
**Status**: PASSED
- `did w` respects configuration (tested both Monday and Sunday)
- `did lw` respects configuration (tested both Monday and Sunday)
- `did stats` respects configuration (verified with Sunday start)
- All commands consistently use the configured week start

### ✅ 4. Headers show the correct date range based on configuration
**Status**: PASSED
- Monday start: "Entries for this week (Jan 5 - Jan 11, 2026):"
- Sunday start: "Entries for this week (Jan 4 - Jan 10, 2026):"
- Last week also shows correct date ranges
- Format is clear and readable

### ✅ 5. Works correctly across year boundaries
**Status**: PASSED
- Monday start last week: Dec 29, 2025 - Jan 4, 2026 ✓
- Sunday start last week: Dec 28, 2025 - Jan 3, 2026 ✓
- Both correctly span the 2025-2026 boundary
- Date formatting handles year transition properly

## Summary

**All 5 acceptance criteria PASSED** ✅

The configurable week start day feature is fully functional:
- Backward compatible (default = Monday)
- Configuration works for both Monday and Sunday
- All week-related commands respect the setting
- Clear date ranges in output
- Robust year boundary handling
- Good error messages for invalid configuration

Test date: Monday, January 5, 2026
Test environment: Isolated test directory with XDG_CONFIG_HOME override

