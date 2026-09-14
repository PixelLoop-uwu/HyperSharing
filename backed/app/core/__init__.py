"""Core: settings, security, shared exceptions, logging.

This package must not know about the outside world (DB, Redis, HTTP, S3).
Infrastructure integrations live in ``app.infrastructure``.
"""
