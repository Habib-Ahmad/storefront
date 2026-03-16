"use client";

import { motion } from "framer-motion";

/** Isometric-style shopping bag with floating motion */
export function ShoppingBagSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -8, 0] }}
      transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Bag body */}
      <motion.path
        d="M50 80 L100 55 L150 80 L150 155 L100 180 L50 155 Z"
        className="fill-primary/10 stroke-primary/40"
        strokeWidth="2"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 1.5, ease: "easeOut" }}
      />
      {/* Bag front face */}
      <path
        d="M50 80 L100 105 L100 180 L50 155 Z"
        className="fill-primary/15"
      />
      {/* Bag right face */}
      <path
        d="M100 105 L150 80 L150 155 L100 180 Z"
        className="fill-primary/8"
      />
      {/* Bag top */}
      <path
        d="M50 80 L100 55 L150 80 L100 105 Z"
        className="fill-primary/20"
      />
      {/* Handle */}
      <motion.path
        d="M75 80 C75 50, 125 50, 125 80"
        className="stroke-primary/50"
        strokeWidth="3"
        strokeLinecap="round"
        fill="none"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 1, delay: 0.5, ease: "easeOut" }}
      />
      {/* Sparkle dots */}
      <motion.circle
        cx="160" cy="50" r="3"
        className="fill-primary/40"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0 }}
      />
      <motion.circle
        cx="40" cy="65" r="2"
        className="fill-primary/30"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0.7 }}
      />
      <motion.circle
        cx="170" cy="120" r="2.5"
        className="fill-primary/25"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 1.4 }}
      />
    </motion.svg>
  );
}

/** Isometric lock / shield with pulse */
export function SecureLockSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -6, 0] }}
      transition={{ duration: 3.5, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Lock body — isometric box */}
      <motion.rect
        x="60" y="95" width="80" height="65" rx="8"
        className="fill-primary/12 stroke-primary/40"
        strokeWidth="2"
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      />
      {/* Lock shackle */}
      <motion.path
        d="M78 95 C78 60, 122 60, 122 95"
        className="stroke-primary/50"
        strokeWidth="4"
        strokeLinecap="round"
        fill="none"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 0.8, delay: 0.3, ease: "easeOut" }}
      />
      {/* Keyhole */}
      <circle cx="100" cy="122" r="8" className="fill-primary/30" />
      <rect x="97" y="125" width="6" height="14" rx="3" className="fill-primary/30" />
      {/* Pulse ring */}
      <motion.circle
        cx="100" cy="127"
        r="30"
        className="stroke-primary/20"
        strokeWidth="2"
        fill="none"
        animate={{ r: [30, 45], opacity: [0.4, 0] }}
        transition={{ duration: 2, repeat: Infinity, ease: "easeOut" }}
      />
    </motion.svg>
  );
}

/** Envelope with letter sliding out */
export function MailSentSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -6, 0] }}
      transition={{ duration: 3.5, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Envelope body */}
      <rect x="40" y="75" width="120" height="80" rx="8" className="fill-primary/12 stroke-primary/40" strokeWidth="2" />
      {/* Envelope flap */}
      <motion.path
        d="M40 75 L100 120 L160 75"
        className="stroke-primary/40"
        strokeWidth="2"
        strokeLinejoin="round"
        fill="none"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 0.8, ease: "easeOut" }}
      />
      {/* Letter sliding up */}
      <motion.rect
        x="55" y="65" width="90" height="60" rx="4"
        className="fill-primary/8 stroke-primary/25"
        strokeWidth="1.5"
        animate={{ y: [65, 50, 65] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      />
      {/* Letter lines */}
      <motion.line x1="70" y1="80" x2="130" y2="80" className="stroke-primary/20" strokeWidth="2" strokeLinecap="round"
        animate={{ y: [0, -15, 0] }} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      />
      <motion.line x1="70" y1="90" x2="115" y2="90" className="stroke-primary/15" strokeWidth="2" strokeLinecap="round"
        animate={{ y: [0, -15, 0] }} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      />
      {/* Sparkle */}
      <motion.circle
        cx="165" cy="65" r="3"
        className="fill-primary/30"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
      />
    </motion.svg>
  );
}

/** Isometric key turning */
export function KeySvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -6, 0] }}
      transition={{ duration: 3.5, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Key head — circle */}
      <motion.circle
        cx="80" cy="90" r="28"
        className="fill-primary/10 stroke-primary/40"
        strokeWidth="2"
        initial={{ scale: 0, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ duration: 0.5 }}
      />
      {/* Key hole in head */}
      <circle cx="80" cy="90" r="10" className="fill-background stroke-primary/30" strokeWidth="2" />
      {/* Key shaft */}
      <motion.line
        x1="108" y1="90" x2="160" y2="90"
        className="stroke-primary/40"
        strokeWidth="4"
        strokeLinecap="round"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 0.6, delay: 0.3 }}
      />
      {/* Key teeth */}
      <motion.path
        d="M145 90 L145 105 M155 90 L155 100"
        className="stroke-primary/40"
        strokeWidth="4"
        strokeLinecap="round"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 0.4, delay: 0.6 }}
      />
      {/* Rotating sparkle */}
      <motion.g
        animate={{ rotate: 360 }}
        transition={{ duration: 8, repeat: Infinity, ease: "linear" }}
        style={{ transformOrigin: "80px 90px" }}
      >
        <circle cx="80" cy="52" r="2.5" className="fill-primary/25" />
        <circle cx="118" cy="90" r="2" className="fill-primary/20" />
      </motion.g>
    </motion.svg>
  );
}

/* ── Dashboard illustrations ── */

/** Rising bar chart with animated bars */
export function ChartRiseSvg({ className }: { className?: string }) {
  const bars = [
    { x: 35, h: 40, delay: 0 },
    { x: 65, h: 65, delay: 0.15 },
    { x: 95, h: 50, delay: 0.3 },
    { x: 125, h: 80, delay: 0.45 },
    { x: 155, h: 95, delay: 0.6 },
  ];
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -5, 0] }}
      transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Grid lines */}
      {[140, 115, 90, 65].map((y) => (
        <line key={y} x1="25" y1={y} x2="175" y2={y} className="stroke-primary/8" strokeWidth="1" />
      ))}
      {/* Bars */}
      {bars.map((bar) => (
        <motion.rect
          key={bar.x}
          x={bar.x}
          width="20"
          rx="4"
          className="fill-primary/20"
          initial={{ y: 165, height: 0 }}
          animate={{ y: 165 - bar.h, height: bar.h }}
          transition={{ duration: 0.8, delay: bar.delay, ease: "easeOut" }}
        />
      ))}
      {/* Trend line */}
      <motion.path
        d="M45 130 L75 105 L105 118 L135 88 L165 72"
        className="stroke-primary/50"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 1.2, delay: 0.8, ease: "easeOut" }}
      />
      {/* Dot at peak */}
      <motion.circle
        cx="165" cy="72" r="4"
        className="fill-primary/60"
        initial={{ scale: 0 }}
        animate={{ scale: [0, 1.3, 1] }}
        transition={{ duration: 0.5, delay: 1.8 }}
      />
      {/* Pulse on dot */}
      <motion.circle
        cx="165" cy="72" r="4"
        className="stroke-primary/30"
        strokeWidth="2"
        fill="none"
        animate={{ r: [4, 14], opacity: [0.5, 0] }}
        transition={{ duration: 2, repeat: Infinity, delay: 2.2 }}
      />
      {/* Arrow up */}
      <motion.path
        d="M165 68 L165 50 M159 56 L165 50 L171 56"
        className="stroke-primary/40"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 2 }}
      />
      {/* Sparkles */}
      <motion.circle cx="30" cy="60" r="2" className="fill-chart-4/40"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2.5, repeat: Infinity }}
      />
      <motion.circle cx="175" cy="45" r="2.5" className="fill-chart-5/35"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.2, 0.5] }}
        transition={{ duration: 2.5, repeat: Infinity, delay: 1 }}
      />
    </motion.svg>
  );
}

/** Open box with items floating out */
export function OpenBoxSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -5, 0] }}
      transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Box body */}
      <motion.path
        d="M45 100 L100 75 L155 100 L155 155 L100 180 L45 155 Z"
        className="fill-primary/10 stroke-primary/35"
        strokeWidth="2"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: 1.2, ease: "easeOut" }}
      />
      {/* Left face */}
      <path d="M45 100 L100 125 L100 180 L45 155 Z" className="fill-primary/15" />
      {/* Right face */}
      <path d="M100 125 L155 100 L155 155 L100 180 Z" className="fill-primary/8" />
      {/* Open flaps */}
      <motion.path
        d="M45 100 L70 75 L100 90 L100 75"
        className="fill-chart-4/10 stroke-primary/30"
        strokeWidth="1.5"
        initial={{ rotate: 0 }}
        animate={{ rotate: [-5, 5, -5] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
        style={{ transformOrigin: "45px 100px" }}
      />
      <motion.path
        d="M155 100 L130 75 L100 90 L100 75"
        className="fill-chart-5/10 stroke-primary/30"
        strokeWidth="1.5"
        initial={{ rotate: 0 }}
        animate={{ rotate: [5, -5, 5] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
        style={{ transformOrigin: "155px 100px" }}
      />
      {/* Floating items */}
      <motion.circle cx="85" cy="70" r="6" className="fill-chart-4/25 stroke-chart-4/40" strokeWidth="1.5"
        animate={{ y: [0, -15, 0], opacity: [0.5, 1, 0.5] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      />
      <motion.rect x="106" y="55" width="12" height="12" rx="3" className="fill-primary/20 stroke-primary/35" strokeWidth="1.5"
        animate={{ y: [0, -12, 0], opacity: [0.5, 1, 0.5], rotate: [0, 15, 0] }}
        transition={{ duration: 3.5, repeat: Infinity, ease: "easeInOut", delay: 0.3 }}
      />
      <motion.path d="M92 48 L100 38 L108 48 L100 52 Z" className="fill-chart-5/25 stroke-chart-5/40" strokeWidth="1.5"
        animate={{ y: [0, -10, 0], opacity: [0.4, 1, 0.4] }}
        transition={{ duration: 2.8, repeat: Infinity, ease: "easeInOut", delay: 0.6 }}
      />
      {/* Sparkles */}
      <motion.circle cx="60" cy="55" r="2" className="fill-primary/30"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
      />
      <motion.circle cx="145" cy="60" r="2.5" className="fill-chart-4/30"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 1.2 }}
      />
    </motion.svg>
  );
}

/** Receipt / order list with check marks */
export function ReceiptSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -5, 0] }}
      transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Paper */}
      <motion.rect
        x="55" y="30" width="90" height="130" rx="8"
        className="fill-primary/8 stroke-primary/35"
        strokeWidth="2"
        initial={{ scaleY: 0 }}
        animate={{ scaleY: 1 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        style={{ transformOrigin: "100px 160px" }}
      />
      {/* Torn bottom edge */}
      <path d="M55 160 L65 155 L75 162 L85 155 L95 162 L105 155 L115 162 L125 155 L135 162 L145 155"
        className="stroke-primary/25" strokeWidth="2" strokeLinecap="round" fill="none" />
      {/* Lines */}
      {[55, 75, 95, 115, 135].map((y, i) => (
        <motion.g key={y}>
          <motion.line
            x1="75" y1={y} x2="130" y2={y}
            className="stroke-primary/15"
            strokeWidth="2"
            strokeLinecap="round"
            initial={{ pathLength: 0 }}
            animate={{ pathLength: 1 }}
            transition={{ duration: 0.4, delay: 0.3 + i * 0.12 }}
          />
          {i < 3 && (
            <motion.path
              d={`M68 ${y - 2} L71 ${y + 1} L76 ${y - 4}`}
              className="stroke-chart-4/50"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              fill="none"
              initial={{ pathLength: 0 }}
              animate={{ pathLength: 1 }}
              transition={{ duration: 0.3, delay: 0.6 + i * 0.15 }}
            />
          )}
        </motion.g>
      ))}
      {/* Dollar sign floating */}
      <motion.text
        x="100" y="185" textAnchor="middle"
        className="fill-primary/25"
        fontSize="16" fontWeight="bold"
        animate={{ y: [185, 175, 185], opacity: [0.3, 0.6, 0.3] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      >₦</motion.text>
      {/* Sparkle */}
      <motion.circle cx="150" cy="40" r="2.5" className="fill-chart-5/35"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0.8 }}
      />
    </motion.svg>
  );
}

/** Delivery truck driving */
export function DeliveryTruckSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {/* Road */}
      <line x1="20" y1="150" x2="180" y2="150" className="stroke-primary/15" strokeWidth="2" />
      {/* Truck body */}
      <motion.g
        animate={{ x: [0, 3, 0, -3, 0] }}
        transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
      >
        {/* Cargo */}
        <rect x="45" y="95" width="75" height="45" rx="4" className="fill-primary/12 stroke-primary/35" strokeWidth="2" />
        {/* Package inside */}
        <rect x="60" y="108" width="18" height="18" rx="3" className="fill-chart-4/15 stroke-chart-4/30" strokeWidth="1.5" />
        <rect x="85" y="108" width="18" height="18" rx="3" className="fill-chart-5/15 stroke-chart-5/30" strokeWidth="1.5" />
        {/* Cab */}
        <path d="M120 105 L150 105 L155 125 L155 140 L120 140 Z" className="fill-primary/15 stroke-primary/35" strokeWidth="2" strokeLinejoin="round" />
        {/* Window */}
        <rect x="128" y="110" width="20" height="14" rx="3" className="fill-primary/8 stroke-primary/25" strokeWidth="1.5" />
      </motion.g>
      {/* Wheels */}
      <motion.g
        animate={{ rotate: 360 }}
        transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
        style={{ transformOrigin: "70px 150px" }}
      >
        <circle cx="70" cy="150" r="10" className="fill-primary/15 stroke-primary/40" strokeWidth="2" />
        <circle cx="70" cy="150" r="3" className="fill-primary/30" />
      </motion.g>
      <motion.g
        animate={{ rotate: 360 }}
        transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
        style={{ transformOrigin: "145px 150px" }}
      >
        <circle cx="145" cy="150" r="10" className="fill-primary/15 stroke-primary/40" strokeWidth="2" />
        <circle cx="145" cy="150" r="3" className="fill-primary/30" />
      </motion.g>
      {/* Speed lines */}
      <motion.line x1="30" y1="120" x2="15" y2="120" className="stroke-primary/20" strokeWidth="2" strokeLinecap="round"
        animate={{ opacity: [0, 0.5, 0], x1: [30, 20, 30] }}
        transition={{ duration: 1.5, repeat: Infinity }}
      />
      <motion.line x1="35" y1="130" x2="18" y2="130" className="stroke-primary/15" strokeWidth="2" strokeLinecap="round"
        animate={{ opacity: [0, 0.5, 0], x1: [35, 25, 35] }}
        transition={{ duration: 1.5, repeat: Infinity, delay: 0.3 }}
      />
      {/* Sparkle */}
      <motion.circle cx="170" cy="85" r="2.5" className="fill-chart-4/35"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
      />
    </motion.svg>
  );
}

/** Wallet with coins */
export function WalletCoinsSvg({ className }: { className?: string }) {
  return (
    <motion.svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      animate={{ y: [0, -5, 0] }}
      transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
    >
      {/* Wallet body */}
      <motion.rect
        x="35" y="75" width="120" height="80" rx="12"
        className="fill-primary/10 stroke-primary/35"
        strokeWidth="2"
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ duration: 0.5 }}
      />
      {/* Wallet flap */}
      <path d="M35 95 L155 95" className="stroke-primary/20" strokeWidth="1.5" />
      {/* Clasp */}
      <rect x="140" y="108" width="25" height="18" rx="9" className="fill-primary/15 stroke-primary/30" strokeWidth="2" />
      <circle cx="152" cy="117" r="4" className="fill-primary/25" />
      {/* Coins floating */}
      <motion.g animate={{ y: [0, -8, 0], rotate: [0, 10, 0] }} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}>
        <circle cx="80" cy="60" r="14" className="fill-chart-5/20 stroke-chart-5/40" strokeWidth="1.5" />
        <text x="80" y="65" textAnchor="middle" className="fill-chart-5/60" fontSize="14" fontWeight="bold">₦</text>
      </motion.g>
      <motion.g animate={{ y: [0, -10, 0], rotate: [0, -8, 0] }} transition={{ duration: 3.2, repeat: Infinity, ease: "easeInOut", delay: 0.4 }}>
        <circle cx="110" cy="52" r="11" className="fill-chart-4/20 stroke-chart-4/35" strokeWidth="1.5" />
        <text x="110" y="57" textAnchor="middle" className="fill-chart-4/55" fontSize="11" fontWeight="bold">₦</text>
      </motion.g>
      <motion.g animate={{ y: [0, -6, 0] }} transition={{ duration: 2.8, repeat: Infinity, ease: "easeInOut", delay: 0.8 }}>
        <circle cx="55" cy="55" r="9" className="fill-primary/15 stroke-primary/30" strokeWidth="1.5" />
        <text x="55" y="59" textAnchor="middle" className="fill-primary/45" fontSize="9" fontWeight="bold">₦</text>
      </motion.g>
      {/* Sparkles */}
      <motion.circle cx="165" cy="70" r="2" className="fill-chart-5/40"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity }}
      />
      <motion.circle cx="35" cy="68" r="2.5" className="fill-chart-4/30"
        animate={{ opacity: [0, 1, 0], scale: [0.5, 1.3, 0.5] }}
        transition={{ duration: 2, repeat: Infinity, delay: 1 }}
      />
    </motion.svg>
  );
}
