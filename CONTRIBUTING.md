# Contributing to MCPTools

Thank you for considering contributing to MCPTools! This document provides guidelines and instructions for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/mcptools.git`
3. Create a new branch: `git checkout -b feature/your-feature-name`

## Development Setup

1. No need to pre-install Go - the setup process will automatically install it if needed
2. Run `make setup` to set up the development environment (this will check/install Go and set up everything else)
3. Make your changes
4. Test your local changes by running `make build` first, then `./bin/mcp [command]`
5. Run tests using `make test`
6. Run the linter using `make lint`

## Pull Request Process

1. Update the README.md with details of changes if needed
2. Follow the existing code style and formatting
3. Add tests for new features
4. Ensure all tests pass and the linter shows no errors
5. Update documentation as needed

## Commit Messages

- Use clear and meaningful commit messages
- Start with a verb in present tense (e.g., "Add feature" not "Added feature")
- Reference issue numbers if applicable

## Code Style

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

## Questions or Problems?

Feel free to open an issue for any questions or problems you encounter.

## License

By contributing, you agree that your contributions will be licensed under the same terms as the main project.
