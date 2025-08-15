"""
FlexFlag Django Integration

Middleware and utilities for integrating FlexFlag with Django applications.
"""

import asyncio
from typing import Optional, Dict, Any
from asgiref.sync import sync_to_async

try:
    from django.conf import settings
    from django.core.exceptions import ImproperlyConfigured
    from django.utils.deprecation import MiddlewareMixin
    DJANGO_AVAILABLE = True
except ImportError:
    DJANGO_AVAILABLE = False

from ..client import FlexFlagClient
from ..types import FlexFlagConfig, EvaluationContext


# Global client instance
_client_instance: Optional[FlexFlagClient] = None


def get_flexflag_client() -> FlexFlagClient:
    """Get the global FlexFlag client instance"""
    global _client_instance
    
    if not DJANGO_AVAILABLE:
        raise ImportError("Django is not installed. Install with: pip install Django>=4.0")
    
    if _client_instance is None:
        if not hasattr(settings, 'FLEXFLAG_CONFIG'):
            raise ImproperlyConfigured(
                "FLEXFLAG_CONFIG must be defined in Django settings"
            )
        
        config_dict = settings.FLEXFLAG_CONFIG
        config = FlexFlagConfig(**config_dict)
        _client_instance = FlexFlagClient(config)
        
        # Initialize the client asynchronously
        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                # Create a task to initialize later
                asyncio.create_task(_client_instance.wait_for_ready())
            else:
                loop.run_until_complete(_client_instance.wait_for_ready())
        except RuntimeError:
            # No event loop, create one
            asyncio.run(_client_instance.wait_for_ready())
    
    return _client_instance


class FlexFlagMiddleware(MiddlewareMixin):
    """Django middleware for FlexFlag integration"""
    
    def __init__(self, get_response):
        self.get_response = get_response
        self.client = get_flexflag_client()
        super().__init__(get_response)
    
    def process_request(self, request):
        """Add FlexFlag client and context to request"""
        request.flexflag = self.client
        
        # Build evaluation context from request
        context = self._build_context_from_request(request)
        request.flag_context = context
        
        return None
    
    def _build_context_from_request(self, request) -> EvaluationContext:
        """Build evaluation context from Django request"""
        user_id = None
        attributes = {}
        
        # Extract user information if available
        if hasattr(request, 'user') and request.user.is_authenticated:
            user_id = str(request.user.id)
            attributes.update({
                'is_staff': getattr(request.user, 'is_staff', False),
                'is_superuser': getattr(request.user, 'is_superuser', False),
                'username': getattr(request.user, 'username', ''),
                'email': getattr(request.user, 'email', ''),
            })
            
            # Add custom user attributes if available
            if hasattr(request.user, 'get_flag_attributes'):
                custom_attrs = request.user.get_flag_attributes()
                if isinstance(custom_attrs, dict):
                    attributes.update(custom_attrs)
        
        # Add request metadata
        attributes.update({
            'path': request.path,
            'method': request.method,
            'user_agent': request.META.get('HTTP_USER_AGENT', ''),
            'ip_address': self._get_client_ip(request),
            'is_secure': request.is_secure(),
        })
        
        # Add session information if available
        session_data = {}
        if hasattr(request, 'session'):
            session_data = {
                'session_key': request.session.session_key,
                'is_empty': request.session.is_empty(),
            }
        
        return EvaluationContext(
            user_id=user_id,
            attributes=attributes,
            session=session_data
        )
    
    def _get_client_ip(self, request) -> str:
        """Get client IP address from request"""
        x_forwarded_for = request.META.get('HTTP_X_FORWARDED_FOR')
        if x_forwarded_for:
            ip = x_forwarded_for.split(',')[0]
        else:
            ip = request.META.get('REMOTE_ADDR')
        return ip or ''


# Template tags for Django templates
if DJANGO_AVAILABLE:
    from django import template
    from django.utils.safestring import mark_safe
    
    register = template.Library()
    
    @register.simple_tag(takes_context=True)
    def feature_flag(context, flag_key, default_value=False):
        """Template tag to evaluate feature flags"""
        request = context.get('request')
        if not request or not hasattr(request, 'flexflag'):
            return default_value
        
        try:
            # Use sync_to_async to handle the async evaluation
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            result = loop.run_until_complete(
                request.flexflag.evaluate(
                    flag_key, 
                    context=getattr(request, 'flag_context', None),
                    default_value=default_value
                )
            )
            loop.close()
            return result
        except Exception:
            return default_value
    
    @register.inclusion_tag('flexflag/feature_gate.html', takes_context=True)
    def feature_gate(context, flag_key, default_enabled=False):
        """Template tag for conditional content rendering"""
        enabled = feature_flag(context, flag_key, default_enabled)
        return {
            'enabled': enabled,
            'flag_key': flag_key
        }
    
    @register.filter
    def if_flag_enabled(content, flag_key):
        """Template filter for conditional content"""
        # This would need request context, simplified version
        return content if flag_key else ""


# Django management command for FlexFlag operations
if DJANGO_AVAILABLE:
    from django.core.management.base import BaseCommand
    
    class Command(BaseCommand):
        """Django management command for FlexFlag operations"""
        help = "FlexFlag management operations"
        
        def add_arguments(self, parser):
            parser.add_argument(
                '--clear-cache',
                action='store_true',
                help='Clear FlexFlag cache'
            )
            parser.add_argument(
                '--test-connection',
                action='store_true', 
                help='Test FlexFlag API connection'
            )
            parser.add_argument(
                '--metrics',
                action='store_true',
                help='Show FlexFlag SDK metrics'
            )
        
        def handle(self, *args, **options):
            client = get_flexflag_client()
            
            if options['clear_cache']:
                asyncio.run(client.clear_cache())
                self.stdout.write(
                    self.style.SUCCESS('FlexFlag cache cleared successfully')
                )
            
            if options['test_connection']:
                try:
                    asyncio.run(client._test_connection())
                    self.stdout.write(
                        self.style.SUCCESS('FlexFlag connection test successful')
                    )
                except Exception as e:
                    self.stdout.write(
                        self.style.ERROR(f'FlexFlag connection test failed: {e}')
                    )
            
            if options['metrics']:
                metrics = client.get_metrics()
                self.stdout.write("FlexFlag SDK Metrics:")
                self.stdout.write(f"  Evaluations: {metrics.evaluations}")
                self.stdout.write(f"  Cache hits: {metrics.cache_hits}")
                self.stdout.write(f"  Cache misses: {metrics.cache_misses}")
                cache_hit_rate = (metrics.cache_hits / metrics.evaluations * 100) if metrics.evaluations > 0 else 0
                self.stdout.write(f"  Cache hit rate: {cache_hit_rate:.1f}%")
                self.stdout.write(f"  Average latency: {metrics.average_latency_ms:.2f}ms")
                self.stdout.write(f"  Network requests: {metrics.network_requests}")
                self.stdout.write(f"  Errors: {metrics.errors}")


# Django app configuration
if DJANGO_AVAILABLE:
    from django.apps import AppConfig
    
    class FlexFlagConfig(AppConfig):
        """Django app configuration for FlexFlag"""
        default_auto_field = 'django.db.models.BigAutoField'
        name = 'flexflag.integrations.django'
        verbose_name = 'FlexFlag'
        
        def ready(self):
            """Initialize FlexFlag when Django app is ready"""
            try:
                get_flexflag_client()
            except Exception as e:
                import logging
                logger = logging.getLogger(__name__)
                logger.error(f"Failed to initialize FlexFlag: {e}")


# Decorators for Django views
def require_feature_flag(flag_key: str, default_enabled: bool = False, redirect_url: str = '/'):
    """Decorator to require a feature flag for view access"""
    def decorator(view_func):
        def wrapper(request, *args, **kwargs):
            if not hasattr(request, 'flexflag'):
                if not default_enabled:
                    from django.shortcuts import redirect
                    return redirect(redirect_url)
                return view_func(request, *args, **kwargs)
            
            try:
                loop = asyncio.new_event_loop()
                asyncio.set_event_loop(loop)
                enabled = loop.run_until_complete(
                    request.flexflag.evaluate(
                        flag_key,
                        context=getattr(request, 'flag_context', None),
                        default_value=default_enabled
                    )
                )
                loop.close()
                
                if enabled:
                    return view_func(request, *args, **kwargs)
                else:
                    from django.shortcuts import redirect
                    return redirect(redirect_url)
                    
            except Exception:
                if default_enabled:
                    return view_func(request, *args, **kwargs)
                else:
                    from django.shortcuts import redirect
                    return redirect(redirect_url)
        
        return wrapper
    return decorator