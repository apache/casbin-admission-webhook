# Contributing to Casbin Admission Webhook

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/casbin-admission-webhook.git
   cd casbin-admission-webhook
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Build**
   ```bash
   make build
   ```

## Making Changes

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Write code
   - Add tests
   - Update documentation

3. **Run Tests and Linting**
   ```bash
   make test
   make lint
   ```

4. **Commit Your Changes**
   
   We use [Conventional Commits](https://www.conventionalcommits.org/):
   
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug"
   git commit -m "docs: update README"
   ```
   
   Types:
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation
   - `test`: Adding tests
   - `refactor`: Code refactoring
   - `perf`: Performance improvements
   - `chore`: Maintenance tasks

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

## Testing

- Write unit tests for new functionality
- Ensure all tests pass
- Aim for good code coverage

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable names
- Add comments for complex logic

## Questions?

Open an issue for discussions or questions.
