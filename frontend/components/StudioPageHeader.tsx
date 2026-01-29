"use client";

import { motion } from "framer-motion";

export default function StudioPageHeader() {
  return (
    <motion.header
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="text-center space-y-2"
    >
      <h1 className="text-xl sm:text-2xl lg:text-3xl font-black uppercase tracking-tight text-white">
        Studio
      </h1>
      <p className="text-xs sm:text-sm text-white/50 uppercase tracking-wider max-w-md mx-auto">
        Paste or upload data → choose template → run engine
      </p>
    </motion.header>
  );
}
