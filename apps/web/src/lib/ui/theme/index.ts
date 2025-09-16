import type { ThemeConfig } from 'flowbite-svelte';

export const appTheme: ThemeConfig = {
  // Button theming
  button: {
    base: 'inline-flex items-center justify-center font-medium rounded-lg text-sm px-5 py-2.5 text-center transition-colors duration-200 focus:outline-none focus:ring-4 focus:ring-primary-300 dark:focus:ring-primary-800'
  },

  // Card theming
  card: {
    base: 'bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800'
  },

  // Badge theming
  badge: {
    base: 'inline-flex items-center px-2.5 py-0.5 text-xs font-medium rounded-full'
  },

  // Input theming
  input: {
    base: 'bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-primary-500 focus:border-primary-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500'
  },

  // Navbar theming
  navbar: {
    base: 'bg-white border-gray-200 px-2 py-2.5 dark:border-gray-700 dark:bg-gray-800'
  },

  // Drawer theming
  drawer: {
    base: 'fixed top-0 left-0 z-40 h-screen p-4 overflow-y-auto transition-transform bg-white w-80 dark:bg-gray-800'
  },

  // Modal theming
  modal: {
    base: 'fixed top-0 left-0 right-0 z-50 w-full p-4 overflow-x-hidden overflow-y-auto md:inset-0 h-[calc(100%-1rem)] md:h-full'
  },

  // Dropdown theming
  dropdown: {
    base: 'z-10 bg-white divide-y divide-gray-100 rounded-lg shadow w-44 dark:bg-gray-700 dark:divide-gray-600'
  },

  // Pagination theming
  pagination: {
    base: 'flex items-center -space-x-px'
  },

  // Alert theming
  alert: {
    base: 'p-4 mb-4 text-sm rounded-lg'
  },

  // Spinner theming
  spinner: {
    base: 'inline animate-spin text-gray-200 dark:text-gray-600'
  },

	rating: {
		base: 'inline-flex items-center justify-center font-medium rounded-full'
	}
};

// Export individual component themes for granular usage
export const buttonTheme = appTheme.button;
export const cardTheme = appTheme.card;
export const badgeTheme = appTheme.badge;
export const inputTheme = appTheme.input;
export const navbarTheme = appTheme.navbar;
export const drawerTheme = appTheme.drawer;
export const modalTheme = appTheme.modal;
export const dropdownTheme = appTheme.dropdown;
export const paginationTheme = appTheme.pagination;
export const alertTheme = appTheme.alert;
export const spinnerTheme = appTheme.spinner;
export const ratingTheme = appTheme.rating;

// Custom component themes
export const restaurantCardTheme = {
  base: 'overflow-hidden rounded-2xl border-0 bg-white shadow-xl transition-all duration-300 hover:scale-[1.01] hover:shadow-2xl dark:bg-gray-800' +
		' dark:shadow-gray-900/50'
};

export const searchInputTheme = {
  base: 'w-full bg-[url(\'/GoogleMaps_Logo_Gray.svg\')] bg-[length:60px_60px] bg-[position:calc(100%-2.25rem)_50%] bg-no-repeat pr-10'
};

export const pageHeaderTheme = {
  base: 'text-primary-800 dark:text-primary-200 text-2xl font-semibold'
};

export const sectionTheme = {
  base: 'space-y-6'
};

export const containerTheme = {
  base: 'container mx-auto max-w-6xl space-y-8 p-6'
};

export const hrTheme = 'my-6 border-gray-200 dark:border-gray-700';

export const tableTheme = 'w-full text-sm text-left text-gray-500 dark:text-gray-400';
