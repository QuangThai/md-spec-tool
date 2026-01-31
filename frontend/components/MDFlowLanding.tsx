"use client";

import { motion, useScroll, useTransform } from "framer-motion";
import {
  ArrowRight,
  ChevronRight,
  Cpu,
  Database,
  Shield,
  Terminal,
  TrendingUp,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import dynamic from "next/dynamic";
import { useRef } from "react";

const LandingHeroPlayer = dynamic(
  () =>
    import("./remotion/LandingHeroPlayer").then(
      (mod) => mod.LandingHeroPlayer
    ),
  {
    ssr: false,
    loading: () => (
      <div className="h-full w-full rounded-xl bg-white/5 animate-pulse" />
    ),
  }
);

const fadeInUp: any = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6, ease: [0.22, 1, 0.36, 1] },
};

const fadeInUpSoft: any = {
  initial: { opacity: 0, y: 12 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true, amount: 0.15 },
  transition: { duration: 0.85, ease: [0.25, 0.46, 0.45, 0.94] },
};

const stagger: any = {
  whileInView: {
    transition: {
      staggerChildren: 0.12,
      delayChildren: 0.04,
    },
  },
};

const features = [
  {
    title: "Precision Engine",
    desc: "Advanced Excel/CSV parsing with intelligent header detection and column mapping.",
    icon: <Cpu className="w-5 h-5" />,
    tag: "Core",
    docId: "parsing",
  },
  {
    title: "Sheet Aware",
    desc: "Native XLSX support with multi-sheet detection and structured table handling.",
    icon: <Database className="w-5 h-5" />,
    tag: "Parsing",
    docId: "ingestion",
  },
  {
    title: "Audit Ready",
    desc: "Clean, standardized Markdown specs optimized for review and version control.",
    icon: <Shield className="w-5 h-5" />,
    tag: "Output",
    docId: "validation",
  },
  {
    title: "Logic Injection",
    desc: "Apply templates to turn raw tables into structured documentation.",
    icon: <Terminal className="w-5 h-5" />,
    tag: "Process",
    docId: "templates",
  },
];

const technicalSteps = [
  {
    id: "01",
    title: "Stream Ingestion",
    desc: "Paste table data or upload Excel/CSV files directly into the engine.",
  },
  {
    id: "02",
    title: "Node Mapping",
    desc: "The parser detects headers and maps columns to spec fields.",
  },
  {
    id: "03",
    title: "Serialization",
    desc: "Automated conversion into standard Markdown specifications.",
  },
];

export default function MDFlowLanding() {
  const containerRef = useRef<HTMLDivElement>(null);
  const router = useRouter();
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  const heroY = useTransform(scrollYProgress, [0, 0.3], [0, -50]);
  const opacity = useTransform(scrollYProgress, [0, 0.2], [1, 0]);

  return (
    <div
      ref={containerRef}
      className="flex flex-col gap-16 sm:gap-24 lg:gap-12"
    >
      {/* 🚀 Hero Section - The Grand Entrance */}
      <section className="relative min-h-[70vh] sm:min-h-[80vh] lg:min-h-[90vh] flex flex-col items-center justify-center px-4 overflow-hidden pt-8 sm:pt-12">
        {/* Advanced Background Gradients - Liquid Glass Influence */}
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[1000px] h-[1000px] bg-accent-orange/10 blur-[150px] rounded-full -z-10 animate-pulse-soft" />
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-accent-gold/5 blur-[120px] rounded-full -z-10" />

        <motion.div
          style={{ y: heroY, opacity }}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, ease: [0.16, 1, 0.3, 1] }}
          className="relative z-10 max-w-5xl mx-auto text-center space-y-6 sm:space-y-8 lg:space-y-12"
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.2 }}
            className="inline-flex items-center gap-2 sm:gap-3 px-3 py-1.5 sm:px-5 sm:py-2.5 rounded-full bg-white/5 border border-white/10 backdrop-blur-xl shadow-2xl"
          >
            <span className="flex h-1.5 w-1.5 sm:h-2 sm:w-2 rounded-full bg-accent-orange shadow-[0_0_10px_rgba(242,123,47,0.8)] animate-pulse" />
            <span className="text-[9px] sm:text-[11px] font-black uppercase tracking-[0.3em] sm:tracking-[0.4em] text-white/90">
              Technical Spec Studio
            </span>
          </motion.div>

          <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl xl:text-8xl 2xl:text-9xl font-black tracking-tighter leading-[0.9] sm:leading-[0.85] text-white overflow-hidden">
            <motion.span
              className="inline-flex flex-wrap justify-center gap-x-[0.15em] gap-y-0"
              initial="hidden"
              animate="visible"
              variants={{
                visible: {
                  transition: {
                    staggerChildren: 0.06,
                    delayChildren: 0.15,
                  },
                },
                hidden: {},
              }}
            >
              {["Automate", "Your"].map((word, i) => (
                <motion.span
                  key={i}
                  className="inline-block"
                  variants={{
                    hidden: { opacity: 0, y: 18 },
                    visible: {
                      opacity: 1,
                      y: 0,
                      transition: {
                        type: "spring",
                        stiffness: 380,
                        damping: 22,
                      },
                    },
                  }}
                >
                  {word}{" "}
                </motion.span>
              ))}
              <br />
              <motion.span
                className="inline-block text-transparent bg-clip-text bg-linear-to-r from-accent-orange via-accent-gold to-white"
                variants={{
                  hidden: { opacity: 0, y: 18 },
                  visible: {
                    opacity: 1,
                    y: 0,
                    transition: {
                      type: "spring",
                      stiffness: 380,
                      damping: 22,
                      delay: 0.12,
                    },
                  },
                }}
              >
                Technical{" "}
              </motion.span>
              <motion.span
                className="inline-block text-transparent bg-clip-text bg-linear-to-r from-accent-orange via-accent-gold to-white"
                variants={{
                  hidden: { opacity: 0, y: 18 },
                  visible: {
                    opacity: 1,
                    y: 0,
                    transition: {
                      type: "spring",
                      stiffness: 380,
                      damping: 22,
                    },
                  },
                }}
              >
                Flow.
              </motion.span>
            </motion.span>
          </h1>

          <p className="max-w-2xl mx-auto text-sm sm:text-base md:text-lg lg:text-2xl text-muted leading-relaxed font-medium">
            Turn complex spreadsheets into clear, standardized technical
            specifications with automated parsing and formatting.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-6 pt-6 sm:pt-10 lg:pt-12">
            <Link href="/studio">
              <motion.button
                whileHover={{ scale: 1.05, y: -2 }}
                whileTap={{ scale: 0.98 }}
                className="btn-primary px-6 sm:px-10 lg:px-12 h-11 sm:h-14 lg:h-16 text-sm sm:text-base group relative overflow-hidden cursor-pointer"
              >
                <span className="relative z-10 flex items-center gap-2 sm:gap-3">
                  Launch Studio
                  <ArrowRight className="w-4 h-4 sm:w-5 sm:h-5 transition-transform group-hover:translate-x-1" />
                </span>
                <div className="absolute inset-0 bg-linear-to-r from-white/0 via-white/20 to-white/0 -translate-x-full group-hover:translate-x-full transition-transform duration-1000" />
              </motion.button>
            </Link>
            <Link href="#features">
              <motion.button
                whileHover={{ backgroundColor: "rgba(255,255,255,0.05)" }}
                className="btn-ghost px-6 sm:px-10 lg:px-12 h-11 sm:h-14 lg:h-16 text-sm sm:text-base border-white/10 group cursor-pointer"
              >
                View Capabilities
                <ChevronRight className="ml-2 w-4 h-4 opacity-0 group-hover:opacity-100 -translate-x-2 group-hover:translate-x-0 transition-all" />
              </motion.button>
            </Link>
          </div>
        </motion.div>

        {/* Floating Technical HUD - Engineering Preview */}
        <motion.div
          initial={{ opacity: 0, y: 100 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6, duration: 1.2 }}
          className="mt-12 sm:mt-16 lg:mt-24 relative w-full max-w-7xl aspect-21/9 px-3 sm:px-5 lg:px-8 rounded-xl sm:rounded-2xl lg:rounded-3xl border border-white/10 bg-black/40 backdrop-blur-3xl shadow-3xl overflow-hidden group"
        >
          <div className="absolute top-0 left-0 w-full h-px bg-linear-to-r from-transparent via-accent-orange/50 to-transparent" />
          <div className="absolute inset-0 bg-linear-to-b from-accent-orange/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />

          <div className="absolute inset-0 p-2 sm:p-4 lg:p-6">
            <LandingHeroPlayer />
          </div>

          {/* Corner Elements */}
          <div className="absolute bottom-4 right-6 text-[9px] font-black text-white/20 uppercase tracking-[0.5em] font-mono">
            TRK ID: 0xFD29A
          </div>
        </motion.div>
      </section>

      {/* 🛠 Feature Grid - Bento Box Style */}
      <section id="features" className="app-container py-6 sm:py-10 lg:py-12">
        <motion.div
          {...fadeInUpSoft}
          className="text-center max-w-4xl mx-auto mb-6 sm:mb-12 lg:mb-16 space-y-3 sm:space-y-5"
        >
          <div className="pill border-accent-orange/30 text-accent-orange mx-auto">
            Core Toolchain
          </div>
          <h2 className="text-3xl sm:text-5xl lg:text-7xl font-black text-white tracking-tighter uppercase leading-[0.9]">
            Engineered for <br />
            <span className="text-accent-orange">Total Control.</span>
          </h2>
          <p className="text-muted text-sm sm:text-base lg:text-xl max-w-2xl mx-auto">
            A focused documentation workflow for engineering teams and technical
            leads.
          </p>
        </motion.div>

        <motion.div
          variants={stagger}
          initial="initial"
          whileInView="whileInView"
          viewport={{ once: true, amount: 0.12 }}
          className="w-full grid grid-cols-2 md:grid-cols-4 gap-2 sm:gap-3 lg:gap-4"
        >
          {features.map((feature, i) => (
            <motion.div
              key={i}
              variants={fadeInUpSoft}
              transition={{ duration: 0.85, ease: [0.25, 0.46, 0.45, 0.94] }}
              whileHover={{ y: -4 }}
              className="bento-card group p-1 relative overflow-hidden min-h-[200px] sm:min-h-[260px] md:min-h-[280px] lg:h-[320px]"
            >
              <div className="absolute inset-0 bg-linear-to-br from-white/5 to-transparent" />
              <div className="relative px-2.5 sm:px-5 lg:px-8 pt-2.5 sm:pt-5 lg:pt-8 pb-5 sm:pb-8 lg:pb-10 h-full flex flex-col z-10">
                <div className="mb-auto min-w-0">
                  <div className="flex items-center justify-between gap-1 mb-2 sm:mb-4 lg:mb-6">
                    <div className="p-1.5 sm:p-2.5 lg:p-3 rounded-lg sm:rounded-xl lg:rounded-2xl bg-accent-orange/10 border border-accent-orange/20 text-accent-orange shadow-lg shadow-accent-orange/5 group-hover:scale-110 transition-transform duration-500 shrink-0 [&>svg]:w-3 [&>svg]:h-3 sm:[&>svg]:w-4 sm:[&>svg]:h-4 lg:[&>svg]:w-5 lg:[&>svg]:h-5">
                      {feature.icon}
                    </div>
                    <span className="text-[8px] sm:text-[9px] lg:text-[10px] font-black uppercase tracking-widest text-white/20 group-hover:text-accent-orange/40 transition-colors shrink-0 truncate">
                      {feature.tag}
                    </span>
                  </div>
                  <h3 className="text-[11px] sm:text-base lg:text-xl font-black text-white uppercase tracking-tight mb-1 sm:mb-2 lg:mb-4 line-clamp-2 sm:line-clamp-none">
                    {feature.title}
                  </h3>
                  <p className="text-[10px] sm:text-xs lg:text-[15px] text-muted leading-snug sm:leading-relaxed font-medium line-clamp-2 sm:line-clamp-3 lg:line-clamp-none">
                    {feature.desc}
                  </p>
                </div>

                <div
                  className="mt-2 sm:mt-4 lg:mt-6 flex items-center gap-1 sm:gap-2 text-[8px] sm:text-[9px] lg:text-[10px] font-black text-accent-orange uppercase tracking-widest opacity-70 sm:opacity-0 sm:group-hover:opacity-100 transition-all -translate-x-1 sm:-translate-x-2 sm:group-hover:translate-x-0 cursor-pointer active:opacity-100"
                  onClick={() => {
                    router.push(`/docs?section=${feature.docId}`);
                  }}
                >
                  Learn More{" "}
                  <ChevronRight className="w-2.5 h-2.5 sm:w-3 sm:h-3 shrink-0" />
                </div>
              </div>
            </motion.div>
          ))}
        </motion.div>
      </section>

      {/* 📊 The Pipeline - Kinetic Workflow */}
      <section className="relative py-10 rounded-xl sm:rounded-2xl lg:rounded-3xl overflow-hidden bg-white/1 border-y border-white/5">
        <div className="app-container grid lg:grid-cols-2 gap-12 sm:gap-16 lg:gap-24 items-center">
          <motion.div {...fadeInUp} className="space-y-8 sm:space-y-12">
            <div className="pill border-accent-gold/30 text-accent-gold">
              Operational Pipeline
            </div>
            <h2 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-black text-white leading-[0.9] tracking-tighter">
              A Zero-Friction <br />
              <span className="text-accent-orange">Technical Path.</span>
            </h2>

            <div className="space-y-8 sm:space-y-12 relative">
              {/* Connector Line */}
              <div className="absolute left-4 sm:left-6 top-6 sm:top-8 bottom-6 sm:bottom-8 w-px bg-linear-to-b from-accent-orange via-accent-orange/60 to-transparent" />

              {technicalSteps.map((step, i) => (
                <div key={i} className="flex gap-4 sm:gap-6 lg:gap-8 group">
                  <div className="flex h-10 w-10 sm:h-12 sm:w-12 shrink-0 items-center justify-center rounded-xl sm:rounded-2xl bg-black border border-accent-orange/30 text-accent-orange font-mono text-xs sm:text-sm font-black z-10 group-hover:border-accent-orange group-hover:shadow-[0_0_15px_rgba(242,123,47,0.35)] transition-all">
                    {step.id}
                  </div>
                  <div className="space-y-1 sm:space-y-2 pt-1">
                    <h4 className="text-base sm:text-lg lg:text-xl font-black text-white uppercase tracking-tight group-hover:text-accent-orange transition-colors">
                      {step.title}
                    </h4>
                    <p className="text-muted text-xs sm:text-sm lg:text-base leading-relaxed">
                      {step.desc}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          {/* Interactive Studio Mockup */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            className="relative"
          >
            <div className="absolute -inset-10 bg-accent-orange/10 blur-[100px] rounded-full animate-pulse-soft" />
            <div className="group surface p-0! overflow-hidden border-white/10 shadow-3xl relative z-10 ring-1 ring-white/10">
              <div className="bg-white/5 border-b border-white/10 p-3 sm:p-5 flex items-center justify-between">
                <div className="flex gap-2 sm:gap-2.5">
                  <div className="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full bg-accent-red/40" />
                  <div className="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full bg-accent-gold/40" />
                  <div className="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full bg-accent-orange/40" />
                </div>
                <div className="flex items-center gap-4 sm:gap-6">
                  <div className="h-1 sm:h-1.5 w-16 sm:w-24 bg-white/10 rounded-full" />
                  <div className="h-1 sm:h-1.5 w-8 sm:w-12 bg-accent-orange/20 rounded-full" />
                </div>
              </div>

              <div className="p-4 sm:p-6 lg:p-10 space-y-4 sm:space-y-6 lg:space-y-8 min-h-[280px] sm:h-[320px] lg:h-[400px]">
                {/* Mock Content Rows */}
                {[1, 2, 3, 4].map((n) => (
                  <motion.div
                    key={n}
                    initial={{ opacity: 0, x: -10 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    transition={{ delay: n * 0.1 }}
                    className="flex items-center gap-3 sm:gap-4 lg:gap-6"
                  >
                    <div className="h-8 w-8 sm:h-10 sm:w-10 rounded-lg sm:rounded-xl bg-white/5 shrink-0" />
                    <div className="flex-1 space-y-3">
                      <div
                        className={`h-2.5 bg-white/5 rounded-full`}
                        style={{ width: `${70 + (n % 3) * 10}%` }}
                      />
                      <div
                        className={`h-2 bg-white/2 rounded-full`}
                        style={{ width: `${40 + (n % 2) * 15}%` }}
                      />
                    </div>
                    <div className="h-6 w-16 rounded-lg bg-accent-orange/5 border border-accent-orange/10" />
                  </motion.div>
                ))}

                {/* Hover Floating Stat - Left */}
                <div className="absolute left-[5%] sm:left-[10%] top-1/2 -translate-y-1/2 w-[140px] sm:w-[180px] lg:w-[200px] p-3 sm:p-4 lg:p-6 rounded-xl sm:rounded-2xl bg-black/85 backdrop-blur-2xl border border-accent-orange/30 shadow-2xl shadow-accent-orange/10 space-y-2 sm:space-y-4 -rotate-3 group-hover:rotate-0 transition-all duration-300 z-20 hidden sm:block">
                  <div className="flex items-center justify-between gap-1.5 sm:gap-2">
                    <div className="flex h-7 w-7 sm:h-9 sm:w-9 shrink-0 items-center justify-center rounded-lg sm:rounded-xl bg-accent-orange/15 border border-accent-orange/25 text-accent-orange">
                      <TrendingUp className="w-3 h-3 sm:w-4 sm:h-4" />
                    </div>
                    <span className="text-[8px] sm:text-[9px] font-black text-accent-orange/80 uppercase tracking-[0.2em]">
                      Efficiency Output
                    </span>
                  </div>
                  <div className="space-y-0.5 sm:space-y-1">
                    <p className="text-xl sm:text-2xl lg:text-3xl font-black text-white tabular-nums leading-none">
                      +84%
                    </p>
                    <p className="text-[8px] sm:text-[9px] text-white/50 font-medium uppercase tracking-wider">
                      time saved vs manual spec
                    </p>
                  </div>
                  <div className="h-1.5 w-full rounded-full bg-white/10 overflow-hidden">
                    <motion.div
                      className="h-full rounded-full bg-linear-to-r from-accent-orange to-accent-gold"
                      initial={{ width: 0 }}
                      whileInView={{ width: "84%" }}
                      viewport={{ once: true }}
                      transition={{
                        duration: 0.8,
                        delay: 0.2,
                        ease: [0.22, 1, 0.36, 1],
                      }}
                    />
                  </div>
                </div>

                {/* Hover Floating Stat - Right */}
                <div className="absolute right-[5%] sm:right-[10%] top-1/2 -translate-y-1/2 w-[140px] sm:w-[180px] lg:w-[200px] p-3 sm:p-4 lg:p-6 rounded-xl sm:rounded-2xl bg-black/85 backdrop-blur-2xl border border-accent-gold/30 shadow-2xl shadow-accent-gold/10 space-y-2 sm:space-y-4 rotate-3 group-hover:rotate-0 transition-all duration-300 z-20 hidden sm:block">
                  <div className="flex items-center justify-between gap-1.5 sm:gap-2">
                    <div className="flex h-7 w-7 sm:h-9 sm:w-9 shrink-0 items-center justify-center rounded-lg sm:rounded-xl bg-accent-gold/15 border border-accent-gold/25 text-accent-gold">
                      <Zap className="w-3 h-3 sm:w-4 sm:h-4" />
                    </div>
                    <span className="text-[8px] sm:text-[9px] font-black text-accent-gold/80 uppercase tracking-[0.2em]">
                      Parse Speed
                    </span>
                  </div>
                  <div className="space-y-0.5 sm:space-y-1">
                    <p className="text-xl sm:text-2xl lg:text-3xl font-black text-white tabular-nums leading-none">
                      12×
                    </p>
                    <p className="text-[8px] sm:text-[9px] text-white/50 font-medium uppercase tracking-wider">
                      faster than manual mapping
                    </p>
                  </div>
                  <div className="h-1.5 w-full rounded-full bg-white/10 overflow-hidden">
                    <motion.div
                      className="h-full rounded-full bg-linear-to-r from-accent-gold to-accent-orange"
                      initial={{ width: 0 }}
                      whileInView={{ width: "100%" }}
                      viewport={{ once: true }}
                      transition={{
                        duration: 0.8,
                        delay: 0.35,
                        ease: [0.22, 1, 0.36, 1],
                      }}
                    />
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>
    </div>
  );
}
