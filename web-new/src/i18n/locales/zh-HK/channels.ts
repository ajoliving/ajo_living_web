/*
 * 頻道頁繁體中文文案。
 * 1. 提供家具與超市優惠頁面文字。
 * 2. 保持前台正式、直接的展示語氣。
 */
export default {
  furniture: {
    eyebrow: 'AJO Living',
    title: '家具',
    subtitle: '集中查看家居傢俱、家庭電器、電子產品與 BB 用品帖子。',
    searchPlaceholder: '搜尋家具帖子',
    allFurniture: '全部',
    homeFurniture: '家居傢俱',
    results: '{count} 件家具帖子',
    empty: '暫時未有家具帖子。',
    loadError: '無法載入家具帖子。',
    browseAll: '查看全部二手帖子',
  },
  offers: {
    eyebrow: 'AJO Living',
    title: '超市優惠',
    subtitle: '集中展示屋苑住戶常用的超市與日用品優惠入口。',
    status: '頻道準備中',
    primaryAction: '前往二手鄰里交易',
    cards: [
      {
        badge: '日用品',
        title: '家居清潔用品',
        description: '預留超市清潔用品、紙品與消耗品優惠位置。',
        discount: '優惠資料待接入',
        expiry: 'AJO Living',
      },
      {
        badge: '食品',
        title: '凍肉與熟食',
        description: '預留社區超市食品與住戶團購資訊位置。',
        discount: '優惠資料待接入',
        expiry: 'AJO Living',
      },
      {
        badge: '會員',
        title: '住戶會員優惠',
        description: '後續可按大廈、單位與會員身份控制可見範圍。',
        discount: '權限模型待接入',
        expiry: 'AJO Living',
      },
      {
        badge: '服務',
        title: '派送與自取',
        description: '預留配送方式、有效期與商戶聯絡資訊。',
        discount: '履約資料待接入',
        expiry: 'AJO Living',
      },
    ],
  },
};
