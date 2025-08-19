// Component-specific theme exports for granular usage
export { buttonTheme } from './index';
export { cardTheme } from './index';
export { badgeTheme } from './index';
export { inputTheme } from './index';
export { navbarTheme } from './index';
export { drawerTheme } from './index';
export { hrTheme } from './index';
export { modalTheme } from './index';
export { dropdownTheme } from './index';
export { tableTheme } from './index';
export { paginationTheme } from './index';
export { alertTheme } from './index';
export { spinnerTheme } from './index';

// Custom component themes that extend the base Flowbite themes
export const restaurantCardTheme = {
  base: 'overflow-hidden rounded-2xl border-0 bg-white shadow-xl transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl dark:bg-gray-800 dark:shadow-gray-900/50'
};

export const searchInputTheme = {
  base: 'w-full bg-[url(\'/GoogleMaps_Logo_Gray.svg\')] bg-[length:60px_60px] bg-[position:calc(100%-2.25rem)_50%] bg-no-repeat pr-10'
};

export const ratingTheme = {
  base: 'rounded-xl bg-primary-100 p-4 dark:bg-primary-900/20',
  text: 'text-lg font-bold text-primary-800 dark:text-primary-200',
  badge: 'rounded-full px-4 py-2 font-bold'
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