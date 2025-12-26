# AI Governance Documentation

This directory contains the immutable contract and technical invariants for the OpenTrusty project. These documents serve as the primary source of truth for all AI agents working on this codebase.

## Mandatory Reading Order

1. **`AI_CONTRACT.md`** (Project Root) - The limits of your agency.
2. **`invariants.md`** - Security and logic rules that MUST NOT be broken.
3. **`authority-model.md`** - Who can do what.
4. **`architecture-map.md`** - Where things live.
5. **`update-matrix.md`** - What to update when you change code.

## Critical Rule

> **If you modify any item listed in `update-matrix.md`, you MUST update the corresponding documentation.**

Failures to strictly flow this rule constitutes a violation of the AI Contract.
