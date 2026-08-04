"""LLM client for agentspec.

Provides model resolution, platform-aware Anthropic client creation,
retry with exponential backoff, and prompt caching.
"""

from __future__ import annotations

import asyncio
import copy
import logging
import os
from collections.abc import Callable, Coroutine
from dataclasses import dataclass
from enum import StrEnum
from typing import Any

import anthropic
from anthropic import APIStatusError, RateLimitError

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Model registry
# ---------------------------------------------------------------------------


class ModelTier(StrEnum):
    SIMPLE = "SIMPLE"
    STANDARD = "STANDARD"
    ADVANCED = "ADVANCED"


@dataclass(frozen=True)
class ModelEntry:
    model_id: str
    tier: ModelTier
    variant: str | None = None


MODEL_REGISTRY: dict[str, ModelEntry] = {
    "claude-haiku-4-5": ModelEntry(
        "claude-haiku-4-5", ModelTier.SIMPLE, variant="standard"
    ),
    "claude-sonnet-4-6": ModelEntry(
        "claude-sonnet-4-6", ModelTier.STANDARD, variant="standard"
    ),
    "claude-opus-4-6": ModelEntry(
        "claude-opus-4-6", ModelTier.ADVANCED, variant="standard"
    ),
}

TIER_DEFAULTS: dict[ModelTier, str] = {
    ModelTier.SIMPLE: "claude-haiku-4-5",
    ModelTier.STANDARD: "claude-sonnet-4-6",
    ModelTier.ADVANCED: "claude-opus-4-6",
}


class CachePolicy(StrEnum):
    NONE = "none"
    DEFAULT = "default"
    EXTENDED = "extended"


def resolve_model(name: str) -> str:
    """Resolve a tier name or model ID to a model ID string."""
    try:
        tier = ModelTier(name)
    except ValueError:
        tier = None

    if tier is not None:
        return TIER_DEFAULTS[tier]

    if name in MODEL_REGISTRY:
        return name

    from agentspec.errors import ConfigError

    valid_options = sorted(MODEL_REGISTRY.keys())
    raise ConfigError(
        f"Unknown model '{name}'. Valid options: {', '.join(valid_options)}"
    )


# ---------------------------------------------------------------------------
# Prompt caching
# ---------------------------------------------------------------------------

_CACHE_TOKEN_THRESHOLDS: dict[str, int] = {
    "claude-sonnet-4-6": 2048,
    "claude-opus-4-6": 4096,
    "claude-haiku-4-5": 4096,
}
_DEFAULT_THRESHOLD: int = 4096

_CACHE_CONTROL: dict[CachePolicy, dict[str, Any] | None] = {
    CachePolicy.NONE: None,
    CachePolicy.DEFAULT: {"type": "ephemeral"},
    CachePolicy.EXTENDED: {"type": "ephemeral", "ttl": "1h"},
}


def _inject_cache_control(
    system: str | list[dict[str, Any]] | None,
    *,
    model: str,
    cache_policy: CachePolicy,
) -> str | list[dict[str, Any]] | None:
    if cache_policy is CachePolicy.NONE or system is None:
        return system

    cache_control = _CACHE_CONTROL[cache_policy]

    if isinstance(system, str):
        total_len = len(system)
    else:
        total_len = sum(
            len(block.get("text", "")) if isinstance(block, dict) else 0
            for block in system
        )

    threshold = _CACHE_TOKEN_THRESHOLDS.get(model, _DEFAULT_THRESHOLD)
    if (total_len // 4) < threshold:
        return system

    if isinstance(system, str):
        blocks: list[dict[str, Any]] = [{"type": "text", "text": system}]
    else:
        blocks = [copy.copy(b) for b in system]

    blocks[-1] = {**blocks[-1], "cache_control": cache_control}
    return blocks


# ---------------------------------------------------------------------------
# Retry
# ---------------------------------------------------------------------------

_RETRY_DELAYS: tuple[float, ...] = (2.0, 30.0, 60.0)


def _is_retryable(exc: Exception) -> bool:
    if isinstance(exc, RateLimitError):
        return True
    if isinstance(exc, APIStatusError) and exc.status_code >= 500:
        return True
    if isinstance(exc, OSError):
        return True
    return False


async def _retry_api_call[T](
    fn: Callable[[], Coroutine[object, object, T]],
    *,
    context: str = "API call",
) -> T:
    max_attempts = len(_RETRY_DELAYS) + 1
    for attempt in range(max_attempts):
        try:
            return await fn()
        except (RateLimitError, APIStatusError, OSError) as exc:
            if not _is_retryable(exc) or attempt == max_attempts - 1:
                raise
            delay = _RETRY_DELAYS[attempt]
            logger.warning(
                "%s: transient error (attempt %d/%d), retrying in %.0fs — %s",
                context,
                attempt + 1,
                max_attempts,
                delay,
                exc,
            )
            await asyncio.sleep(delay)
    raise AssertionError("unreachable")  # pragma: no cover


# ---------------------------------------------------------------------------
# Client factory
# ---------------------------------------------------------------------------


def _create_async_client() -> anthropic.AsyncAnthropic:
    if os.environ.get("CLAUDE_CODE_USE_VERTEX") == "1":
        try:
            import google.auth  # noqa: F401
        except ModuleNotFoundError:
            raise RuntimeError(
                "CLAUDE_CODE_USE_VERTEX=1 is set but google-auth is not installed. "
                "Run: pip install 'anthropic[vertex]'"
            ) from None
        from anthropic import AsyncAnthropicVertex

        return AsyncAnthropicVertex()  # type: ignore[return-value]

    if os.environ.get("CLAUDE_CODE_USE_BEDROCK") == "1":
        try:
            import boto3  # type: ignore[import-untyped]  # noqa: F401
        except ModuleNotFoundError:
            raise RuntimeError(
                "CLAUDE_CODE_USE_BEDROCK=1 is set but boto3 is not installed. "
                "Run: pip install 'anthropic[bedrock]'"
            ) from None
        from anthropic import AsyncAnthropicBedrock

        return AsyncAnthropicBedrock()  # type: ignore[return-value]

    return anthropic.AsyncAnthropic()


# ---------------------------------------------------------------------------
# High-level AI call
# ---------------------------------------------------------------------------


async def ai_call(
    *,
    model_tier: str,
    max_tokens: int,
    messages: list[dict[str, Any]],
    system: str | list[dict[str, Any]] | None = None,
    context: str,
    cache_policy: CachePolicy = CachePolicy.DEFAULT,
    **kwargs: Any,
) -> tuple[str | None, Any]:
    """Async AI call: resolve model, create client, retry, extract text.

    Returns (response_text_or_none, raw_response).
    """
    model_id = resolve_model(model_tier)

    modified_system = _inject_cache_control(
        system, model=model_id, cache_policy=cache_policy
    )

    async def _call() -> Any:
        client = _create_async_client()
        call_kwargs: dict[str, Any] = dict(
            model=model_id,
            max_tokens=max_tokens,
            messages=messages,
            **kwargs,
        )
        if modified_system is not None:
            call_kwargs["system"] = modified_system

        try:
            async with client.messages.stream(**call_kwargs) as stream:
                return await stream.get_final_message()
        except anthropic.BadRequestError as exc:
            if "cache_control" in str(exc).lower() and modified_system is not system:
                logger.warning(
                    "cache_control caused API error; retrying without caching"
                )
                fallback_kwargs: dict[str, Any] = dict(
                    model=model_id,
                    max_tokens=max_tokens,
                    messages=messages,
                    **kwargs,
                )
                if system is not None:
                    fallback_kwargs["system"] = system
                async with client.messages.stream(**fallback_kwargs) as stream:
                    return await stream.get_final_message()
            raise
        finally:
            await client.close()

    response = await _retry_api_call(_call, context=context)

    content = getattr(response, "content", None)
    text = getattr(content[0], "text", None) if content else None
    return text, response
