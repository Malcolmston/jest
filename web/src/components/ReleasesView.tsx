import { ReleaseList, ghrepo } from 'go-ui';
import type { RelLib } from 'go-ui';
import { JEST } from '../data';

// Scoped to this repository only: the live release history is read straight from
// the GitHub Releases API for malcolmston/jest.
const RELEASE_LIBS: RelLib[] = [
  { name: JEST.name, icon: JEST.icon, accent: JEST.accent, repo: ghrepo(JEST), url: JEST.repo },
];

// ReleasesView renders the live release-history + changelog tab.
export function ReleasesView() {
  return (
    <section className="view active" id="view-releases">
      <div className="sec-h"><span className="bar" /><h2 style={{ margin: 0 }}>Releases &amp; changelogs</h2></div>
      <p className="muted">jest ships automated semver releases — the moment a <code>VERSION</code> bump lands on <code>main</code>, a tag and GitHub Release are cut and the moving <code>stable</code> tag advances. The list below is read <b>live</b> from the GitHub Releases API, newest first, so it is never out of date. Full history lives in <code>CHANGELOG.md</code>.</p>
      <div style={{ marginTop: '1.4rem' }}><ReleaseList libs={RELEASE_LIBS} /></div>
    </section>
  );
}
