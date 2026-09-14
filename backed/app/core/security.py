"""Security primitives.

Password hashing uses PBKDF2-HMAC-SHA256 from the standard library, so it works
out of the box with no third-party crypto dependency.
"""

import hashlib
import hmac
import secrets

from app.core.config import settings


def hash_password(password: str) -> str:
    """Return ``algorithm$iterations$salt_hex$hash_hex``."""
    salt = secrets.token_bytes(settings.password_salt_bytes)
    digest = hashlib.pbkdf2_hmac(
        "sha256",
        password.encode(),
        salt,
        settings.password_hash_iterations,
        dklen=settings.password_hash_bytes,
    )
    return (
        f"{settings.password_hash_algorithm}${settings.password_hash_iterations}"
        f"${salt.hex()}${digest.hex()}"
    )


def verify_password(password: str, encoded: str) -> bool:
    """Constant-time check of ``password`` against a stored ``encoded`` value."""
    try:
        algorithm, iterations, salt_hex, hash_hex = encoded.split("$")
        if algorithm != settings.password_hash_algorithm:
            return False
        salt = bytes.fromhex(salt_hex)
        expected = bytes.fromhex(hash_hex)
        actual = hashlib.pbkdf2_hmac(
            "sha256", password.encode(), salt, int(iterations), dklen=len(expected)
        )
    except (ValueError, TypeError):
        return False
    return hmac.compare_digest(actual, expected)


def generate_token() -> str:
    """URL-safe random token (sessions, one-time links, etc.)."""
    return secrets.token_urlsafe(32)
