# Overview
Repository notes, so it doens't polute other instructions files

# Notes

## cmd/ Folder Purpose

The `cmd/` directory contains **application entry points** (main packages) for executables.

### Standard Go Layout
- Each subdirectory in `cmd/` represents a **separate executable**
- Each contains its own `main.go` with `package main`

### When to Use cmd/

| Scenario | Use `cmd/` | Use `main.go` at root |
|----------|------------|----------------------|
| Single application | ❌ Optional | ✅ Yes (simpler) |
| Multiple applications | ✅ Yes | ❌ No |
| Library package | ❌ No | ❌ No |
| Quick prototyping | ❌ No | ✅ Yes |
| Production multi-service | ✅ Yes | ❌ No |