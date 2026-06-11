/*
 * Channel page English copy.
 * 1. Provides furniture and supermarket offer page text.
 * 2. Keeps public channel language direct and operational.
 */
export default {
  furniture: {
    eyebrow: 'AJO Living',
    title: 'Furniture',
    subtitle: 'Browse home furniture, office furniture, and home decor listings in one place.',
    searchPlaceholder: 'Search furniture listings',
    allFurniture: 'All furniture',
    homeFurniture: 'Home furniture',
    officeFurniture: 'Office furniture',
    homeDecor: 'Home decor',
    results: '{count} furniture listings',
    empty: 'No furniture listings are available yet.',
    loadError: 'Unable to load furniture listings.',
    browseAll: 'View all secondhand listings',
  },
  offers: {
    eyebrow: 'AJO Living',
    title: 'Supermarket Offers',
    subtitle: 'A structured entry for supermarket and household offers used by residents.',
    status: 'Channel preparing',
    primaryAction: 'Go to Neighbour Marketplace',
    cards: [
      {
        badge: 'Household',
        title: 'Home cleaning goods',
        description: 'Reserved for supermarket cleaning supplies, paper goods, and household consumables.',
        discount: 'Offer data pending',
        expiry: 'AJO Living',
      },
      {
        badge: 'Food',
        title: 'Frozen food and deli',
        description: 'Reserved for community supermarket food and resident group-buy information.',
        discount: 'Offer data pending',
        expiry: 'AJO Living',
      },
      {
        badge: 'Member',
        title: 'Resident member offers',
        description: 'Future visibility can be scoped by building, unit, and member identity.',
        discount: 'Permission model pending',
        expiry: 'AJO Living',
      },
      {
        badge: 'Service',
        title: 'Delivery and pickup',
        description: 'Reserved for fulfilment method, validity period, and merchant contact information.',
        discount: 'Fulfilment data pending',
        expiry: 'AJO Living',
      },
    ],
  },
};
