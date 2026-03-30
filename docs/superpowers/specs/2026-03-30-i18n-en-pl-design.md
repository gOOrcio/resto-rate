# i18n: English + Polish Locale Support

**Date:** 2026-03-30
**Status:** Approved
**Branch:** `feat/i18n-en-pl`

---

## Summary

Add bilingual (English / Polish) support to the SvelteKit frontend. Locale is stored as a user preference (`default_language` on the user profile), synced via a `PARAGLIDE_LOCALE` cookie, and toggled from an EN / PL segmented control in Profile → Preferences. All UI chrome strings are extracted to Paraglide message files. User-generated content (restaurant names, review text, tag labels from the DB) is not translated.

---

## Architecture

### 1. Proto change

Add `default_language` to `UpdateMyProfileRequest` (auth service proto):

```protobuf
message UpdateMyProfileRequest {
  string username = 1;
  string default_region = 2;         // reserved/unused
  bool is_dark_mode_enabled = 3;
  bool set_is_dark_mode_enabled = 4;
  string default_language = 5;        // NEW — "en" or "pl", empty = no change
}
```

`UserProto` already carries `default_language` (field 5). No model change needed; the `users.default_language` column already exists.

### 2. Go auth service

In `UpdateMyProfile` handler: when `req.Msg.DefaultLanguage` is non-empty, write it to the user record. Validate it is one of `{"en", "pl"}` and return `CodeInvalidArgument` otherwise.

### 3. Frontend locale module — `src/lib/state/locale.svelte.ts`

Thin wrapper over Paraglide:

```ts
import { setLocale, getLocale } from '$lib/paraglide/runtime';

export type Locale = 'en' | 'pl';
const VALID: Locale[] = ['en', 'pl'];

function createLocale() {
  let current = $state<Locale>(getLocale() as Locale);
  return {
    get current() { return current; },
    set(l: Locale) {
      if (!VALID.includes(l)) return;
      setLocale(l);
      current = l;
    }
  };
}

export const locale = createLocale();
```

### 4. Auth integration (`src/lib/state/auth.svelte.ts`)

In `setUser(u)`: when `u` is non-null and `u.defaultLanguage` is a valid locale, call `locale.set(u.defaultLanguage as Locale)`. This runs on every login and on `GetCurrentUser` at startup, so the locale cookie is always in sync with the saved preference.

### 5. Profile page — locale switcher

In the Preferences section, below dark mode. EN / PL segmented control:
- Optimistic: `locale.set()` immediately on click
- Async: `client.auth.updateMyProfile({ defaultLanguage })` in background
- Revert on failure (same pattern as dark mode toggle)

### 6. Message files

Flat JSON, one file per locale. Key naming: `section_description`. See the full translation table below.

---

## Scope: What Is and Is Not Translated

**Translated (UI chrome):** All static labels, buttons, headings, placeholders, error messages, empty-state text, ARIA labels.

**Not translated (user-generated / external data):**
- Restaurant names, addresses, cities, countries
- Review comments and dish highlights written by users
- Tag labels — these come from the database (`tags.label` column). Tags are English-only for now. A future migration can add a `label_pl` column if needed.
- Google Places data (names, phone numbers, website URLs, hours text — these come pre-localised from Google's API based on the `languageCode` parameter, which we already set from `navigator.language`)

---

## Message Keys & Translations

### `common_*` — shared across multiple pages

| Key | English | Polish |
|-----|---------|--------|
| `common_save` | Save | Zapisz |
| `common_saving` | Saving… | Zapisywanie… |
| `common_cancel` | Cancel | Anuluj |
| `common_loading` | Loading… | Ładowanie… |
| `common_delete` | Delete | Usuń |
| `common_deleting` | Deleting… | Usuwanie… |
| `common_edit` | Edit | Edytuj |
| `common_back` | Back | Wróć |
| `common_retry` | Retry | Spróbuj ponownie |
| `common_clear` | Clear | Wyczyść |
| `common_clear_all` | Clear all | Wyczyść wszystko |
| `common_clear_filters` | Clear filters | Wyczyść filtry |
| `common_details_and_reviews` | Details and reviews | Szczegóły i opinie |
| `common_highlights` | Highlights: | Wyróżnione dania: |
| `common_would_visit_again_yes` | Would visit again | Wrócę tu |
| `common_would_visit_again_maybe` | Maybe again | Może wrócę |
| `common_would_visit_again_no` | Wouldn't return | Nie wrócę |
| `common_sort` | Sort | Sortuj |
| `common_sort_newest` | Newest first | Od najnowszych |
| `common_sort_oldest` | Oldest first | Od najstarszych |
| `common_sort_rating_high` | Highest rated | Najwyżej oceniane |
| `common_sort_rating_low` | Lowest rated | Najniżej oceniane |
| `common_sort_name_az` | Name A–Z | Nazwa A–Z |
| `common_sort_name_za` | Name Z–A | Nazwa Z–A |
| `common_filter_tags` | Tags | Tagi |
| `common_filter_rating` | Rating | Ocena |
| `common_filter_city` | City | Miasto |
| `common_filter_country` | Country | Kraj |
| `common_filter_all_cities` | All cities | Wszystkie miasta |
| `common_filter_all_countries` | All countries | Wszystkie kraje |
| `common_filter_comment` | Comment contains | Komentarz zawiera |
| `common_filter_comment_placeholder` | Search in comments… | Szukaj w komentarzach… |
| `common_filter_min_rating` | Min ★ | Min ★ |
| `common_filter_max_rating` | Max ★ | Maks ★ |
| `common_filter_rating_to` | to | do |
| `common_filter_rating_error` | Min rating cannot exceed max rating | Minimalna ocena nie może przekraczać maksymalnej |
| `common_filter_tag_mode_match` | Match: | Dopasuj: |
| `common_filter_tag_mode_any` | Any | Dowolne |
| `common_filter_tag_mode_all` | All | Wszystkie |
| `common_removing` | Removing… | Usuwanie… |

### `nav_*` — navigation / header

| Key | English | Polish |
|-----|---------|--------|
| `nav_brand` | Restorate | Restorate |
| `nav_my_reviews` | My Reviews | Moje opinie |
| `nav_wishlist` | Wishlist | Lista życzeń |
| `nav_friends` | Friends | Znajomi |
| `nav_my_profile` | My Profile | Mój profil |
| `nav_sign_in` | Sign in | Zaloguj się |
| `nav_sign_out` | Sign out | Wyloguj się |
| `nav_account` | Account | Konto |
| `nav_preferences` | Preferences | Preferencje |
| `nav_light_mode` | Light mode | Tryb jasny |
| `nav_dark_mode` | Dark mode | Tryb ciemny |
| `nav_sign_in_to_restorate` | Sign in to Restorate | Zaloguj się do Restorate |
| `nav_find_friend` | Find a Friend | Znajdź znajomego |

### `home_*` — landing page

| Key | English | Polish |
|-----|---------|--------|
| `home_hero_title` | Your personal restaurant diary | Twój osobisty dziennik restauracji |
| `home_hero_subtitle` | Track every meal, build your wishlist, and discover where your friends love to eat. | Oceniaj posiłki, buduj listę życzeń i odkrywaj ulubione miejsca znajomych. |
| `home_get_started` | Get started — it's free | Zacznij — to bezpłatne |
| `home_already_account` | Already have an account? | Masz już konto? |
| `home_feature_review_title` | Rate & Review | Oceniaj i recenzuj |
| `home_feature_review_desc` | Leave personal ratings and notes for every restaurant you visit. Add tags like Romantic, Business lunch, or Great value. | Dodawaj osobiste oceny i notatki do każdej odwiedzonej restauracji. Używaj tagów takich jak Romantyczna, Biznesowy lunch czy Świetna wartość. |
| `home_feature_wishlist_title` | Wishlist | Lista życzeń |
| `home_feature_wishlist_desc` | Save restaurants you want to try. They're automatically removed when you leave a review. | Zapisuj restauracje, które chcesz odwiedzić. Są automatycznie usuwane po dodaniu opinii. |
| `home_feature_friends_title` | Friends | Znajomi |
| `home_feature_friends_desc` | Connect with friends and see their reviews when you search. Discover hidden gems through people you trust. | Połącz się ze znajomymi i oglądaj ich opinie podczas wyszukiwania. Odkrywaj ukryte perełki dzięki ludziom, którym ufasz. |

### `reviews_*` — My Reviews page

| Key | English | Polish |
|-----|---------|--------|
| `reviews_page_title` | My Reviews | Moje opinie |
| `reviews_add` | + Add review | + Dodaj opinię |
| `reviews_search_label` | Search for a restaurant to review | Wyszukaj restaurację do oceny |
| `reviews_choose_different` | Choose different restaurant | Wybierz inną restaurację |
| `reviews_empty_no_filters` | No reviews yet. Add your first one above. | Brak opinii. Dodaj pierwszą powyżej. |
| `reviews_empty_with_filters` | No reviews match the current filters. | Brak opinii pasujących do filtrów. |

### `wishlist_*` — My Wishlist page

| Key | English | Polish |
|-----|---------|--------|
| `wishlist_page_title` | My Wishlist | Moja lista życzeń |
| `wishlist_add` | + Add place | + Dodaj miejsce |
| `wishlist_search_label` | Search for a restaurant to save | Wyszukaj restaurację do zapisania |
| `wishlist_already_reviewed` | You've already reviewed this place — it can't be added to your wishlist. | Ta restauracja ma już Twoją opinię — nie można jej dodać do listy życzeń. |
| `wishlist_save_error` | Failed to save. Please try again. | Nie udało się zapisać. Spróbuj ponownie. |
| `wishlist_save_btn` | Save to wishlist | Dodaj do listy życzeń |
| `wishlist_rate_instead` | Write a review instead | Napisz opinię zamiast tego |
| `wishlist_rate_place` | Rate this place | Oceń to miejsce |
| `wishlist_save_tags` | Save tags | Zapisz tagi |
| `wishlist_empty_no_filters` | Your wishlist is empty. Add a place to get started. | Twoja lista życzeń jest pusta. Dodaj miejsce, aby zacząć. |
| `wishlist_empty_with_filters` | No wishlist items match the current filters. | Brak elementów listy pasujących do filtrów. |

### `profile_*` — Profile page

| Key | English | Polish |
|-----|---------|--------|
| `profile_title` | My Profile | Mój profil |
| `profile_section_identity` | Identity | Tożsamość |
| `profile_email_label` | Email | E-mail |
| `profile_member_since` | Member since | Członek od |
| `profile_username_label` | Username / handle | Nazwa użytkownika |
| `profile_username_not_set` | Not set | Nie ustawiono |
| `profile_username_placeholder` | e.g. jane_eats | np. jan_smak |
| `profile_username_hint` | 3–30 characters · lowercase letters, digits, underscores | 3–30 znaków · małe litery, cyfry, podkreślniki |
| `profile_username_error` | Username must be 3–30 chars: lowercase letters, digits, underscores only. | Nazwa użytkownika musi mieć 3–30 znaków: małe litery, cyfry i podkreślniki. |
| `profile_username_saved` | Username saved! | Nazwa użytkownika zapisana! |
| `profile_username_save_error` | Failed to save username | Nie udało się zapisać nazwy użytkownika |
| `profile_section_activity` | Activity | Aktywność |
| `profile_section_preferences` | Preferences | Preferencje |
| `profile_dark_mode_label` | Dark mode | Tryb ciemny |
| `profile_dark_mode_desc` | Synced across all your devices | Synchronizowane na wszystkich urządzeniach |
| `profile_locale_label` | Language | Język |
| `profile_locale_en` | English | English |
| `profile_locale_pl` | Polski | Polski |
| `profile_section_danger` | Danger zone | Strefa zagrożenia |
| `profile_sign_out_all_label` | Sign out all devices | Wyloguj ze wszystkich urządzeń |
| `profile_sign_out_all_desc` | Invalidates all active sessions including this one | Unieważnia wszystkie aktywne sesje, w tym bieżącą |
| `profile_sign_out_all_btn` | Sign out all | Wyloguj wszystkich |
| `profile_sign_out_all_busy` | Signing out… | Wylogowywanie… |
| `profile_delete_label` | Delete account | Usuń konto |
| `profile_delete_desc` | Permanently deletes your account, reviews, and wishlist. This cannot be undone. | Trwale usuwa Twoje konto, opinie i listę życzeń. Tej operacji nie można cofnąć. |
| `profile_delete_placeholder` | Type "DELETE" to confirm | Wpisz „DELETE", aby potwierdzić |
| `profile_delete_confirm_error` | Type DELETE to confirm. | Wpisz DELETE, aby potwierdzić. |
| `profile_delete_failed` | Failed to delete account. | Nie udało się usunąć konta. |

> **Note:** The delete confirmation word stays as `DELETE` in both locales to keep the validation logic simple and unambiguous.

### `friends_*` — Friends page

| Key | English | Polish |
|-----|---------|--------|
| `friends_page_title` | Friends | Znajomi |
| `friends_add_label` | Add a friend | Dodaj znajomego |
| `friends_add_placeholder` | Email address or @username | Adres e-mail lub @nazwa_użytkownika |
| `friends_send_request` | Send Request | Wyślij zaproszenie |
| `friends_sending` | Sending… | Wysyłanie… |
| `friends_request_sent` | Friend request sent to {name} | Zaproszenie wysłane do {name} |
| `friends_request_failed` | Failed to send request | Nie udało się wysłać zaproszenia |
| `friends_pending` | Pending requests | Oczekujące zaproszenia |
| `friends_accept` | Accept | Akceptuj |
| `friends_accepting` | Accepting… | Akceptowanie… |
| `friends_decline` | Decline | Odrzuć |
| `friends_declining` | Declining… | Odrzucanie… |
| `friends_my_friends` | My Friends ({count}) | Moi znajomi ({count}) |
| `friends_empty` | No friends yet. Send a request above to get started. | Brak znajomych. Wyślij zaproszenie powyżej, aby zacząć. |
| `friends_view_profile` | View Profile | Zobacz profil |

### `friend_profile_*` — Friend's profile page (`/friends/[userId]`)

| Key | English | Polish |
|-----|---------|--------|
| `friend_profile_title` | Friend's Profile | Profil znajomego |
| `friend_profile_back` | ← Back to friends | ← Wróć do znajomych |
| `friend_profile_not_friends` | You need to be friends to view this profile. | Musisz być znajomym, aby zobaczyć ten profil. |
| `friend_profile_go_friends` | Go to Friends | Idź do znajomych |
| `friend_profile_tab_reviews` | Reviews | Opinie |
| `friend_profile_tab_wishlist` | Wishlist | Lista życzeń |
| `friend_profile_no_reviews` | No reviews yet. | Brak opinii. |
| `friend_profile_no_wishlist` | Empty wishlist. | Pusta lista życzeń. |
| `friend_profile_no_match_reviews` | No reviews match the current filters. | Brak opinii pasujących do filtrów. |
| `friend_profile_no_match_wishlist` | No wishlist items match the current filters. | Brak elementów listy pasujących do filtrów. |
| `friend_profile_see_all_reviews` | See all reviews → | Zobacz wszystkie opinie → |
| `friend_profile_city_placeholder` | e.g. Paris | np. Paryż |
| `friend_profile_country_placeholder` | e.g. France | np. Francja |
| `friend_profile_sort_any_or` | Any (OR) | Dowolny (LUB) |
| `friend_profile_sort_all_and` | All (AND) | Wszystkie (I) |

### `restaurant_*` — Restaurant detail page

| Key | English | Polish |
|-----|---------|--------|
| `restaurant_load_error` | Failed to load restaurant data. | Nie udało się załadować danych restauracji. |
| `restaurant_wishlist_btn` | Wishlist | Lista życzeń |
| `restaurant_wishlisted_btn` | Wishlisted | Na liście życzeń |
| `restaurant_write_review` | Write a review | Napisz opinię |
| `restaurant_your_review` | Your review | Twoja opinia |
| `restaurant_no_reviews_yet` | No reviews yet | Brak opinii |
| `restaurant_reviews_from_friends` | {count} {count, plural, one {review} other {reviews}} from friends | {count} {count, plural, one {opinia} few {opinie} other {opinii}} od znajomych |
| `restaurant_google_label` | Google Places data | Dane Google Places |
| `restaurant_google_loading` | Loading Google details… | Ładowanie danych Google… |
| `restaurant_google_unavailable` | Google details unavailable. | Dane Google niedostępne. |
| `restaurant_google_reviews` | ({count} Google reviews) | ({count} opinii Google) |
| `restaurant_status_open` | Open | Otwarte |
| `restaurant_status_temp_closed` | Temporarily closed | Tymczasowo zamknięte |
| `restaurant_status_perm_closed` | Permanently closed | Trwale zamknięte |
| `restaurant_hours_today` | Today's hours | Godziny otwarcia dziś |
| `restaurant_features` | Features | Udogodnienia |
| `restaurant_feature_dine_in` | Dine-in | Na miejscu |
| `restaurant_feature_takeout` | Takeout | Na wynos |
| `restaurant_feature_delivery` | Delivery | Dostawa |
| `restaurant_feature_outdoor` | Outdoor seating | Ogródek |
| `restaurant_feature_reservations` | Reservations | Rezerwacje |
| `restaurant_open_maps` | Open in Google Maps | Otwórz w Mapach Google |
| `restaurant_friends_reviews` | Friends' reviews ({count}) | Opinie znajomych ({count}) |
| `restaurant_friend_fallback` | Friend | Znajomy |

### `rating_form_*` — RatingForm component

| Key | English | Polish |
|-----|---------|--------|
| `rating_form_editing` | Editing review | Edytowanie opinii |
| `rating_form_rate` | Rate this place | Oceń to miejsce |
| `rating_form_select_star` | Please select a star rating | Wybierz ocenę gwiazdkową |
| `rating_form_add_comment` | Add comment | Dodaj komentarz |
| `rating_form_comment_placeholder` | What did you think? | Co sądzisz? |
| `rating_form_add_tags` | Add tags ({count}) | Dodaj tagi ({count}) |
| `rating_form_add_visit_details` | Add visit details | Dodaj szczegóły wizyty |
| `rating_form_visit_date` | Visit date | Data wizyty |
| `rating_form_visit_again` | Visit again? | Odwiedzić ponownie? |
| `rating_form_visit_again_yes` | Yes | Tak |
| `rating_form_visit_again_maybe` | Maybe | Może |
| `rating_form_visit_again_no` | No | Nie |
| `rating_form_dish_highlights` | Dish highlights | Polecane dania |
| `rating_form_dish_placeholder` | Dishes you'd recommend… | Dania, które polecasz… |
| `rating_form_save_error` | Failed to save review | Nie udało się zapisać opinii |
| `rating_form_update` | Update rating | Zaktualizuj ocenę |
| `rating_form_save` | Save rating | Zapisz ocenę |

### `tag_picker_*` — TagPicker component

| Key | English | Polish |
|-----|---------|--------|
| `tag_picker_loading` | Loading tags… | Ładowanie tagów… |
| `tag_picker_error` | Failed to load tags. | Nie udało się załadować tagów. |
| `tag_picker_empty` | No tags available. | Brak dostępnych tagów. |

### `search_*` — RestaurantSearch component

| Key | English | Polish |
|-----|---------|--------|
| `search_login_required` | Please log in to search restaurants. | Zaloguj się, aby wyszukiwać restauracje. |
| `search_min_chars` | Type at least 2 characters to search… | Wpisz co najmniej 2 znaki, aby wyszukać… |

### `expandable_*` — ExpandableRestaurantInfo component

| Key | English | Polish |
|-----|---------|--------|
| `expandable_show_google` | Show Google details | Pokaż dane Google |
| `expandable_hide_details` | Hide details | Ukryj szczegóły |
| `expandable_load_failed` | Failed to load Google details. | Nie udało się załadować danych Google. |
| `expandable_google_source` | Data from Google Places API | Dane z Google Places API |
| `expandable_open_maps` | Open in Maps | Otwórz w Mapach |

---

## Tag Translation via Slugs

Tag slugs are stable identifiers — they map directly to message keys. No DB schema change is needed.

### Lookup utility — `src/lib/i18n/tags.ts`

```ts
import * as m from '$lib/paraglide/messages';

type MessageKey = keyof typeof m;

export function tagLabel(slug: string): string {
  const key = `tag_${slug.replace(/-/g, '_')}` as MessageKey;
  return typeof m[key] === 'function' ? (m[key] as () => string)() : slug;
}

export function tagCategoryLabel(category: string): string {
  const key = `tag_category_${category.toLowerCase()}` as MessageKey;
  return typeof m[key] === 'function' ? (m[key] as () => string)() : category;
}
```

`TagPicker` and `TagFilter` call `tagLabel(tag.slug)` instead of `tag.label`. The `label` column in the DB remains as the English seed reference but is never used for display.

### `tag_category_*` — category headings (6 keys)

| Key | English | Polish |
|-----|---------|--------|
| `tag_category_cuisine` | Cuisine | Kuchnia |
| `tag_category_vibe` | Vibe | Klimat |
| `tag_category_price` | Price | Cena |
| `tag_category_dietary` | Dietary | Dieta |
| `tag_category_group` | Group | Grupa |
| `tag_category_occasion` | Occasion | Okazja |

### `tag_*` — individual tag labels (36 keys)

| Key | English | Polish |
|-----|---------|--------|
| `tag_italian` | Italian | Włoska |
| `tag_japanese` | Japanese | Japońska |
| `tag_mexican` | Mexican | Meksykańska |
| `tag_chinese` | Chinese | Chińska |
| `tag_indian` | Indian | Indyjska |
| `tag_french` | French | Francuska |
| `tag_thai` | Thai | Tajska |
| `tag_american` | American | Amerykańska |
| `tag_mediterranean` | Mediterranean | Śródziemnomorska |
| `tag_korean` | Korean | Koreańska |
| `tag_romantic` | Romantic | Romantyczna |
| `tag_casual` | Casual | Na luzie |
| `tag_family_friendly` | Family Friendly | Rodzinna |
| `tag_date_night` | Date Night | Randka |
| `tag_business_lunch` | Business Lunch | Biznesowy lunch |
| `tag_lively` | Lively | Tętniąca życiem |
| `tag_quiet` | Quiet | Spokojna |
| `tag_trendy` | Trendy | Modna |
| `tag_budget` | Budget | Budżetowa |
| `tag_mid_range` | Mid-Range | Średnia półka |
| `tag_expensive` | Expensive | Droga |
| `tag_splurge` | Splurge | Na bogato |
| `tag_vegan` | Vegan | Wegańska |
| `tag_vegetarian` | Vegetarian | Wegetariańska |
| `tag_gluten_free` | Gluten-Free | Bezglutenowa |
| `tag_halal` | Halal | Halal |
| `tag_kosher` | Kosher | Koszerna |
| `tag_dairy_free` | Dairy-Free | Bez nabiału |
| `tag_solo` | Solo | Solo |
| `tag_couple` | Couple | Para |
| `tag_small_group` | Small Group | Mała grupa |
| `tag_large_group` | Large Group | Duża grupa |
| `tag_birthday` | Birthday | Urodziny |
| `tag_anniversary` | Anniversary | Rocznica |
| `tag_brunch` | Brunch | Brunch |
| `tag_late_night` | Late Night | Późny wieczór |
| `tag_celebration` | Celebration | Świętowanie |
| `tag_quick_bite` | Quick Bite | Szybki kąsek |

---

## Tests

### `src/lib/i18n/messages.test.ts` (Vitest unit)

```ts
import en from '../../../../messages/en.json';
import pl from '../../../../messages/pl.json';

test('every en key exists in pl with a non-empty value', () => {
  for (const key of Object.keys(en)) {
    if (key === '$schema') continue;
    expect(pl).toHaveProperty(key);
    expect((pl as Record<string, string>)[key]).toBeTruthy();
  }
});

test('pl has no extra keys not present in en', () => {
  for (const key of Object.keys(pl)) {
    if (key === '$schema') continue;
    expect(en).toHaveProperty(key);
  }
});
```

### `src/lib/state/locale.svelte.test.ts` (Vitest browser)

- Mock `$lib/paraglide/runtime` to expose `setLocale` / `getLocale` spies
- Test: `locale.set('pl')` calls `setLocale('pl')` and `locale.current` becomes `'pl'`
- Test: `locale.set('en')` calls `setLocale('en')` and `locale.current` becomes `'en'`
- Test: `locale.set('de')` (invalid) does NOT call `setLocale`, current unchanged

### Existing tests

`src/demo.spec.ts` — no UI strings, unaffected.
`src/routes/page.svelte.test.ts` — checks for an `<h1>`. After i18n the h1 content comes from `m.home_hero_title()`. Update the test to assert the element exists rather than its exact English text, or assert `m.home_hero_title()` directly.

---

## Implementation Order

1. Proto change + codegen (`default_language` in `UpdateMyProfileRequest`)
2. Go: `UpdateMyProfile` handler — save `default_language`
3. Message files — fill `en.json` + `pl.json` with all UI keys + tag keys above
4. `locale.svelte.ts` — locale state module
5. `src/lib/i18n/tags.ts` — `tagLabel()` + `tagCategoryLabel()` utilities
6. Auth integration — `setLocale` on login
7. Profile page — EN/PL segmented control
8. Replace all hardcoded strings in pages + components with `m.key()` calls
9. Update `TagPicker` + `TagFilter` to use `tagLabel()` / `tagCategoryLabel()`
10. Tests — `messages.test.ts` + `locale.svelte.test.ts`
11. Fix existing `page.svelte.test.ts`
12. `bunx nx run api:build` + `bunx nx run web:check` + `bunx nx run web:test`
13. Open PR → Copilot review loop
