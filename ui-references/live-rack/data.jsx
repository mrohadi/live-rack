// Sample data for live-rack prototype
window.LR_DATA = {
  zones: [
    { id: 'A1', name: 'A1 · Receiving',   type: 'receiving',  x: 4,  y: 4,  w: 24, h: 18, items: 142, capacity: 200, sales: 0,    color: '#7c3aed' },
    { id: 'A2', name: 'A2 · Apparel',     type: 'apparel',    x: 30, y: 4,  w: 28, h: 28, items: 412, capacity: 480, sales: 1820, color: '#2563eb' },
    { id: 'A3', name: 'A3 · Footwear',    type: 'apparel',    x: 60, y: 4,  w: 18, h: 28, items: 188, capacity: 220, sales: 940,  color: '#2563eb' },
    { id: 'A4', name: 'A4 · Electronics', type: 'electronic', x: 80, y: 4,  w: 16, h: 28, items: 96,  capacity: 140, sales: 3210, color: '#0891b2' },
    { id: 'B1', name: 'B1 · Home',        type: 'home',       x: 4,  y: 24, w: 24, h: 22, items: 256, capacity: 360, sales: 670,  color: '#16a34a' },
    { id: 'B2', name: 'B2 · Returns',     type: 'returns',    x: 30, y: 34, w: 18, h: 12, items: 38,  capacity: 80,  sales: 0,    color: '#d97706' },
    { id: 'B3', name: 'B3 · Outbound',    type: 'outbound',   x: 50, y: 34, w: 28, h: 12, items: 64,  capacity: 100, sales: 0,    color: '#dc2626' },
    { id: 'B4', name: 'B4 · Cold Storage',type: 'cold',       x: 80, y: 34, w: 16, h: 12, items: 71,  capacity: 100, sales: 0,    color: '#0891b2' },
    { id: 'C1', name: 'C1 · Bulk',        type: 'bulk',       x: 4,  y: 48, w: 44, h: 24, items: 1240,capacity: 2000,sales: 0,    color: '#5b6577' },
    { id: 'C2', name: 'C2 · Showroom',    type: 'showroom',   x: 50, y: 48, w: 30, h: 24, items: 88,  capacity: 120, sales: 5640, color: '#16a34a' },
    { id: 'C3', name: 'C3 · Staging',     type: 'staging',    x: 82, y: 48, w: 14, h: 24, items: 22,  capacity: 60,  sales: 0,    color: '#5b6577' },
  ],
  items: [
    { sku: 'LR-7821', name: 'Merino Crew Sweater · Slate',     zone: 'A2', qty: 24, price: 89,  velocity: 'high',   updated: '2m ago',  status: 'in-stock' },
    { sku: 'LR-9034', name: 'Trail Runner GTX · 9.5',          zone: 'A3', qty: 8,  price: 145, velocity: 'high',   updated: '6m ago',  status: 'low' },
    { sku: 'LR-3318', name: 'Wireless Earbuds · Onyx',         zone: 'A4', qty: 32, price: 199, velocity: 'medium', updated: '12m ago', status: 'in-stock' },
    { sku: 'LR-6611', name: 'Linen Throw · Sand',              zone: 'B1', qty: 14, price: 64,  velocity: 'low',    updated: '38m ago', status: 'in-stock' },
    { sku: 'LR-1240', name: 'Espresso Machine · Steel',        zone: 'C2', qty: 3,  price: 549, velocity: 'medium', updated: '4m ago',  status: 'low' },
    { sku: 'LR-4502', name: 'Yoga Mat · 6mm Sage',             zone: 'B1', qty: 41, price: 38,  velocity: 'medium', updated: '21m ago', status: 'in-stock' },
    { sku: 'LR-8870', name: 'Smart Bulb 4-pack',               zone: 'A4', qty: 0,  price: 42,  velocity: 'high',   updated: 'just now',status: 'out' },
    { sku: 'LR-2204', name: 'Canvas Tote · Olive',             zone: 'A2', qty: 56, price: 28,  velocity: 'low',    updated: '1h ago',  status: 'in-stock' },
    { sku: 'LR-5566', name: 'Cast Iron Skillet · 12"',         zone: 'B1', qty: 17, price: 79,  velocity: 'medium', updated: '46m ago', status: 'in-stock' },
    { sku: 'LR-7711', name: 'Wool Beanie · Charcoal',          zone: 'A2', qty: 22, price: 24,  velocity: 'high',   updated: '8m ago',  status: 'in-stock' },
  ],
  tasks: [
    { id: 'T-204', title: 'Restock A3 footwear from C1-bulk', assignee: 'Maya R.',   priority: 'high',   due: 'Today 4:00pm',   status: 'in-progress', zone: 'A3' },
    { id: 'T-205', title: 'Cycle count electronics A4',       assignee: 'Theo K.',   priority: 'medium', due: 'Today 6:30pm',   status: 'todo',        zone: 'A4' },
    { id: 'T-206', title: 'Receive PO #4421 from supplier',   assignee: 'Jordan P.', priority: 'high',   due: 'Tomorrow 9:00am',status: 'todo',        zone: 'A1' },
    { id: 'T-207', title: 'Photograph showroom restock',      assignee: 'Aniyah B.', priority: 'low',    due: 'Fri',            status: 'todo',        zone: 'C2' },
    { id: 'T-208', title: 'Move slow movers to clearance',    assignee: 'Maya R.',   priority: 'medium', due: 'Fri',            status: 'review',      zone: 'B1' },
    { id: 'T-209', title: 'Repair pallet jack #3',            assignee: 'Theo K.',   priority: 'low',    due: 'Mon',            status: 'done',        zone: 'C1' },
  ],
  pipelines: {
    'item-restoration': {
      stages: ['Intake', 'Triage', 'Repair', 'QA', 'Restocked'],
      cards: [
        { id: 'RS-001', title: 'Espresso Machine · scratched casing',  sku: 'LR-1240', stage: 0, age: '2h',   priority: 'low',    owner: 'JP' },
        { id: 'RS-002', title: 'Earbuds · case hinge loose',           sku: 'LR-3318', stage: 0, age: '5h',   priority: 'medium', owner: 'TK' },
        { id: 'RS-003', title: 'Trail Runner · sole separation',       sku: 'LR-9034', stage: 1, age: '1d',   priority: 'medium', owner: 'JP' },
        { id: 'RS-004', title: 'Linen Throw · dye irregularity',       sku: 'LR-6611', stage: 1, age: '1d',   priority: 'low',    owner: 'AB' },
        { id: 'RS-005', title: 'Cast Iron · re-season',                sku: 'LR-5566', stage: 2, age: '3d',   priority: 'high',   owner: 'MR' },
        { id: 'RS-006', title: 'Smart Bulb 4-pack · packaging tear',   sku: 'LR-8870', stage: 2, age: '12h',  priority: 'medium', owner: 'TK' },
        { id: 'RS-007', title: 'Canvas Tote · strap restitch',         sku: 'LR-2204', stage: 3, age: '4d',   priority: 'low',    owner: 'AB' },
        { id: 'RS-008', title: 'Yoga Mat · re-roll & cert',            sku: 'LR-4502', stage: 4, age: '6d',   priority: 'low',    owner: 'MR' },
      ]
    }
  },
  integrations: [
    { name: 'Shopify',     cat: 'E-commerce',  status: 'connected',  last: '12s ago',  bg: '#16a34a', initials: 'SH' },
    { name: 'Square POS',  cat: 'POS',         status: 'connected',  last: 'live',     bg: '#0b1220', initials: 'SQ' },
    { name: 'Stripe',      cat: 'Payments',    status: 'connected',  last: '1m ago',   bg: '#7c3aed', initials: 'ST' },
    { name: 'Shippo',      cat: 'Shipping',    status: 'connected',  last: '4m ago',   bg: '#0891b2', initials: 'SP' },
    { name: 'Zebra Scanner',cat: 'Hardware',    status: 'connected',  last: 'live',    bg: '#d97706', initials: 'ZB' },
    { name: 'OpenWeather', cat: 'External',    status: 'connected',  last: '5m ago',   bg: '#2563eb', initials: 'OW' },
    { name: 'Transit API', cat: 'External',    status: 'connected',  last: '8m ago',   bg: '#16a34a', initials: 'TR' },
    { name: 'Klaviyo',     cat: 'Marketing',   status: 'paused',     last: '2h ago',   bg: '#dc2626', initials: 'KL' },
    { name: 'NetSuite',    cat: 'ERP',         status: 'available',  last: '—',        bg: '#5b6577', initials: 'NS' },
  ],
  activity: [
    { who: 'Square POS',  text: 'Sold ', em: 'Trail Runner GTX · 9.5', tail: ' · $145.00', t: '14:22', kind: 'success' },
    { who: 'Maya R.',     text: 'Moved 12 units of ', em: 'Merino Crew Sweater', tail: ' to A2', t: '14:19', kind: 'accent' },
    { who: 'Scanner #3',  text: 'Mis-scan blocked: ', em: 'LR-3318', tail: ' rejected from B1', t: '14:14', kind: 'warn' },
    { who: 'Shopify',     text: 'New order ', em: '#10241', tail: ' · 3 items · $267.00', t: '14:09', kind: 'success' },
    { who: 'Theo K.',     text: 'Started cycle count in ', em: 'A4 Electronics', tail: '', t: '14:02', kind: 'accent' },
    { who: 'Sensor',      text: 'Cold Storage ', em: 'B4', tail: ' temperature normal · 38°F', t: '13:58', kind: 'muted' },
    { who: 'OpenWeather', text: 'Rain forecast → boost ', em: 'umbrella SKUs', tail: ' to A2-front', t: '13:50', kind: 'violet' },
  ],
  // 24h sales sparkline (relative)
  sales24: [12,9,6,5,4,3,5,9,16,22,28,34,40,38,42,48,52,55,58,49,40,32,24,18],
  // heatmap 7 days x 24h
  heatmap: Array.from({length:7}, (_,d) => Array.from({length:24},(_,h) => {
    const peak = Math.exp(-Math.pow((h-14)/4,2)) * (d===5||d===6?1.0:0.7);
    const morn = Math.exp(-Math.pow((h-10)/3,2)) * 0.4;
    return Math.min(1, peak + morn + Math.random()*0.08);
  })),
  zonePerf: [
    { zone: 'A2 Apparel',     sales: 18420, change: 12 },
    { zone: 'A4 Electronics', sales: 14080, change: 28 },
    { zone: 'C2 Showroom',    sales: 11560, change: -4 },
    { zone: 'A3 Footwear',    sales:  9410, change:  6 },
    { zone: 'B1 Home',        sales:  6720, change: -2 },
  ],
  combos: [
    { a: 'Wool Beanie',         b: 'Merino Sweater',     lift: '+3.4×' },
    { a: 'Trail Runner GTX',    b: 'Yoga Mat',           lift: '+2.1×' },
    { a: 'Cast Iron Skillet',   b: 'Linen Throw',        lift: '+1.8×' },
    { a: 'Smart Bulb 4-pack',   b: 'Wireless Earbuds',   lift: '+1.6×' },
  ],
  demographics: [
    { label: '18–24', pct: 14 },
    { label: '25–34', pct: 32 },
    { label: '35–44', pct: 26 },
    { label: '45–54', pct: 16 },
    { label: '55+',   pct: 12 },
  ]
};
