"""
FlexFlag Python SDK Setup
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("requirements.txt", "r", encoding="utf-8") as fh:
    requirements = [line.strip() for line in fh if line.strip() and not line.startswith("#")]

setup(
    name="flexflag",
    version="1.0.0",
    author="FlexFlag Team",
    author_email="support@flexflag.io",
    description="FlexFlag Python SDK for feature flag management with intelligent caching",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/flexflag/flexflag",
    project_urls={
        "Bug Tracker": "https://github.com/flexflag/flexflag/issues",
        "Documentation": "https://docs.flexflag.io",
        "Homepage": "https://flexflag.io",
    },
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Internet :: WWW/HTTP",
        "Topic :: System :: Systems Administration",
    ],
    packages=find_packages(exclude=["tests", "tests.*", "examples", "examples.*"]),
    python_requires=">=3.7",
    install_requires=requirements,
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-asyncio>=0.21.0",
            "pytest-cov>=4.0.0",
            "black>=23.0.0",
            "flake8>=6.0.0",
            "mypy>=1.0.0",
            "pre-commit>=3.0.0",
        ],
        "django": [
            "Django>=3.2",
        ],
        "flask": [
            "Flask>=2.0.0",
        ],
        "fastapi": [
            "FastAPI>=0.100.0",
            "uvicorn[standard]>=0.20.0",
        ],
        "redis": [
            "redis>=4.0.0",
        ],
        "async": [
            "aiohttp>=3.8.0",
            "asyncio-mqtt>=0.11.0",
        ],
    },
    keywords=[
        "feature-flags",
        "feature-toggles", 
        "flexflag",
        "ab-testing",
        "rollout",
        "experiments",
        "cache",
        "offline-first",
        "python",
        "django",
        "flask",
        "fastapi"
    ],
    include_package_data=True,
    zip_safe=False,
)