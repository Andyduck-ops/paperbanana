import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Layout } from './Layout';

describe('Layout', () => {
  it('renders children', () => {
    render(
      <Layout>
        <div>Test content</div>
      </Layout>
    );
    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('renders header when provided', () => {
    render(
      <Layout header={<div>Header</div>}>
        <div>Content</div>
      </Layout>
    );
    expect(screen.getByText('Header')).toBeInTheDocument();
  });

  it('renders footer when provided', () => {
    render(
      <Layout footer={<div>Footer</div>}>
        <div>Content</div>
      </Layout>
    );
    expect(screen.getByText('Footer')).toBeInTheDocument();
  });

  it('renders sidebar when provided', () => {
    render(
      <Layout sidebar={<div>Sidebar</div>}>
        <div>Content</div>
      </Layout>
    );
    expect(screen.getByText('Sidebar')).toBeInTheDocument();
  });

  it('has responsive classes', () => {
    const { container } = render(
      <Layout>
        <div>Content</div>
      </Layout>
    );
    const main = container.querySelector('main');
    expect(main?.className).toContain('md:flex-row');
  });
});
