import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DocsView } from '../../../src/components/DocsView';
import type { DocIndex } from 'go-ui';

// A minimal DocIndex the stubbed fetch returns for DocsApp's doc.json request.
const DOC_INDEX: DocIndex = {
  module: 'github.com/malcolmston/jest',
  packages: [
    {
      importPath: 'github.com/malcolmston/jest',
      name: 'jest',
      synopsis: 'Package jest provides a Jest-style assertion and mocking framework over the standard testing package.',
      doc: 'Package jest provides a Jest-style assertion and mocking framework over the standard testing package.',
      consts: [],
      vars: [],
      types: [
        {
          name: 'Mock',
          signature: 'type Mock struct{}',
          doc: 'Mock records the calls made to it and can be configured with canned return values.',
          consts: [],
          vars: [],
          funcs: [],
          methods: [],
        },
      ],
      funcs: [{ name: 'Expect', signature: 'func Expect[T any](t TestReporter, actual T) *Matcher[T]', doc: 'Expect begins a fluent assertion for the given actual value.' }],
    },
  ],
};

describe('DocsView', () => {
  beforeEach(() => {
    // DocsApp fetches doc.json; return the small index.
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      if (String(input).includes('doc.json')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(DOC_INDEX) } as Response);
      }
      return new Promise<Response>(() => {});
    }) as unknown as typeof fetch;
  });

  it('renders the inline React API reference from the fetched doc.json', async () => {
    const { container } = render(<DocsView />);
    expect(container.querySelector('#view-docs')).not.toBeNull();
    expect(
      screen.getByRole('heading', { level: 2, name: /API documentation/ }),
    ).toBeInTheDocument();

    // DocsApp fetches asynchronously, then renders the package view + symbols.
    expect(await screen.findByRole('heading', { name: /package jest/ })).toBeInTheDocument();
    expect(container.querySelector('#sym-Expect'), 'func Expect symbol card').not.toBeNull();
    expect(container.querySelector('#sym-Mock'), 'type Mock symbol card').not.toBeNull();

    // The secondary link to the raw generated static HTML remains.
    expect(screen.getByRole('link', { name: /Open the raw generated HTML/ })).toHaveAttribute('href', './api/');
  });
});
