"""Background jobs (arq worker).

This package owns only the arq integration: Redis settings, ``WorkerSettings``
and the worker lifecycle. Business tasks live in their feature packages
(``features/*/jobs.py``) and are registered in ``worker.py``.
"""

from app.infrastructure.jobs.worker import WorkerSettings

__all__ = ["WorkerSettings"]
