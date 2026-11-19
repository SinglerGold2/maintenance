// The module's own texts, in the language the operator set on the extension's
// settings page. "system" follows the browser; a language without a file falls
// back to English.
import czech from './locales/cs.json';
import german from './locales/de.json';
import english from './locales/en.json';

const books = { en: english, de: german, cs: czech } as const;

export type Tag = keyof typeof books;

const tags = Object.keys(books) as Tag[];

/** The language the page is drawing in. */
let current = $state<Tag>('en');

const spoken = (): string =>
  typeof navigator === 'undefined' ? 'en' : navigator.language;

/** The language the browser asks for, when this build ships it. */
export function systemTag(): Tag {
  const tag = spoken().slice(0, 2).toLowerCase();

  return (tags as string[]).includes(tag) ? (tag as Tag) : 'en';
}

/** Applies the operator's choice: a tag, or "system" for the browser's. */
export function applyLanguage(setting: string | undefined): void {
  if (setting === undefined || setting === '' || setting === 'system') {
    current = systemTag();

    return;
  }

  // A tag this build does not ship falls away like "system" would: the browser
  // decides, and English is the last word.
  current = (tags as string[]).includes(setting)
    ? (setting as Tag)
    : systemTag();
}

/** The tag in use, for a page that carries its own prose per language. */
export function currentTag(): Tag {
  return current;
}

/** Every text the module draws. */
export function t(key: string): string {
  const book = books[current] as Record<string, string>;

  return book[key] ?? (english as Record<string, string>)[key] ?? key;
}
