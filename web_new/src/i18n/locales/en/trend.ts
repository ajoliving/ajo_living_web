/*
 * Market trend English copy.
 * 1. Provides the rental trend header, chart, and regional card text.
 * 2. Provides loading, error, and updated-month text.
 */
export default {
  eyebrow: 'MARKET DATA',
  title: 'Hong Kong Residential Rental Trends',
  subtitle: 'Public Rating and Valuation Department data for the average monthly rent of Class A to C private homes over the past six months.',
  allHongKong: 'All Hong Kong',
  regions: {
    hk: 'Hong Kong Island',
    kln: 'Kowloon',
    nt: 'New Territories',
  },
  loading: 'Loading public rental trends',
  loadError: 'Unable to load public rental trends.',
  retry: 'Reload',
  chartTitle: 'Average rent trend (HK$/sqft/month)',
  updatedThrough: 'Updated through {month}',
  chartAria: 'Hong Kong residential rental trend line chart',
  latestByRegion: 'Latest rent by region',
  pricePerSqft: '{value}/sqft',
  monthly: 'month on month',
  method: 'Simple average of Class A to C private residential rents by region, converted from monthly rent per square metre to monthly rent per square foot.',
};
