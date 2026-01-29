"use client";

import { motion } from "framer-motion";
import {
  ArrowRight,
  CheckCircle2,
  Cpu,
  Database,
  FileCode,
  Layers,
  MousePointer2,
  Search,
  Shield,
  Terminal,
  Zap,
} from "lucide-react";
import Link from "next/link";

const fadeInUp: any = {
  initial: { opacity: 0, y: 40 },
  whileInView: { opacity: 0.99, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.8, ease: [0.16, 1, 0.3, 1] },
};

const stagger = {
  whileInView: {
    transition: {
      staggerChildren: 0.1,
    },
  },
};

const features = [
  {
    title: "Precision Engine",
    desc: "Advanced TSV/CSV parsing with intelligent column mapping and validation.",
    icon: <Cpu className="w-6 h-6" />,
  },
  {
    title: "Sheet Aware",
    desc: "Native support for complex XLSX files with multi-sheet detection logic.",
    icon: <Database className="w-6 h-6" />,
  },
  {
    title: "Audit Ready",
    desc: "Generate clean, standardized Markdown specifications optimized for Git workflows.",
    icon: <Shield className="w-6 h-6" />,
  },
  {
    title: "Logic Injection",
    desc: "Apply various schema templates to transform raw data into structured specs.",
    icon: <Terminal className="w-6 h-6" />,
  },
];

export default function MDFlowLanding() {
  return (
    <div className="flex flex-col gap-32 pb-32">
      {/* 🚀 Hero Section - The Grand Entrance */}
      <section className="relative min-h-[80vh] flex flex-col items-center justify-center text-center px-4 overflow-hidden">
        {/* Background Decorative Elements */}
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-accent-orange/10 blur-[120px] rounded-full -z-10" />
        <div className="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-accent-gold/5 blur-[100px] rounded-full -z-10" />

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 1, ease: [0.16, 1, 0.3, 1] }}
          className="relative z-10 max-w-4xl mx-auto space-y-10"
        >
          <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-white/5 border border-white/10 backdrop-blur-md">
            <span className="flex h-2 w-2 rounded-full bg-accent-orange animate-pulse" />
            <span className="text-[10px] font-bold uppercase tracking-[0.3em] text-white/80">
              Technical Studio v1.2 Release
            </span>
          </div>

          <h1 className="text-6xl sm:text-8xl font-black tracking-tighter leading-[0.9] text-white">
            Automate Your <br />
            <span className="text-transparent bg-clip-text bg-linear-to-r from-accent-orange via-accent-gold to-white">
              Technical Flow.
            </span>
          </h1>

          <p className="max-w-xl mx-auto text-lg sm:text-xl text-muted leading-relaxed font-medium">
            Stop manually maintaining specifications. Convert spreadsheets into
            industry-standard technical streams with precision automation.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-6 pt-10">
            <Link href="/app">
              <motion.button
                whileHover={{ scale: 1.05, y: -2 }}
                whileTap={{ scale: 0.98 }}
                className="btn-primary px-10 h-16 text-base group"
              >
                Launch Studio
                <ArrowRight className="ml-3 w-5 h-5 transition-transform group-hover:translate-x-1" />
              </motion.button>
            </Link>
            <Link href="#features">
              <motion.button
                whileHover={{ backgroundColor: "rgba(255,255,255,0.05)" }}
                className="btn-ghost px-10 h-16 text-base border-white/10"
              >
                Explorer Engine
              </motion.button>
            </Link>
          </div>
        </motion.div>

        {/* Floating Code Snippet / Visual Cue */}
        <motion.div
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 1 }}
          className="mt-20 relative w-full max-w-5xl aspect-video rounded-2xl border border-white/10 bg-black/40 backdrop-blur-2xl shadow-2xl overflow-hidden group"
        >
          <div className="absolute inset-0 bg-linear-to-b from-accent-orange/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />
          <div className="p-8 h-full flex flex-col">
            <div className="flex items-center gap-2 mb-6 opacity-30">
              <div className="w-2.5 h-2.5 rounded-full bg-white/20" />
              <div className="w-2.5 h-2.5 rounded-full bg-white/20" />
              <div className="w-2.5 h-2.5 rounded-full bg-white/20" />
            </div>
            <div className="flex-1 font-mono text-sm sm:text-base text-white/40 overflow-hidden">
              <span className="text-accent-orange">mdflow</span> process
              dev_spec.xlsx <br />
              <span className="text-white/60">
                -- Reading plane: [System_Architecture]
              </span>{" "}
              <br />
              <span className="text-white/60">-- Validating nodes...</span>{" "}
              <br />
              <span className="text-accent-gold">
                -- Generating stream: output.md
              </span>{" "}
              <br />
              <span className="text-accent-green">
                {">> "} SUCCESS: 142 objects mapped.
              </span>
            </div>
          </div>
        </motion.div>
      </section>

      {/* 🛠 Feature Grid - The Core Power */}
      <section id="features" className="app-container py-20">
        <motion.div
          {...fadeInUp}
          className="text-center max-w-3xl mx-auto mb-20 space-y-6"
        >
          <h2 className="text-4xl sm:text-5xl font-black text-white tracking-tighter uppercase">
            Built for <span className="text-accent-orange">Precision.</span>
          </h2>
          <p className="text-muted text-lg font-medium">
            A specialized toolchain designed to handle the complexities of
            engineering documentation at scale.
          </p>
        </motion.div>

        <motion.div
          variants={stagger}
          initial="initial"
          whileInView="whileInView"
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
        >
          {features.map((feature, i) => (
            <motion.div
              key={i}
              variants={fadeInUp}
              whileHover={{ y: -8, backgroundColor: "rgba(255,255,255,0.03)" }}
              className="bento-card group p-10! relative overflow-hidden"
            >
              <div className="absolute top-0 right-0 p-4 opacity-0 group-hover:opacity-100 transition-opacity">
                <Zap className="w-4 h-4 text-accent-orange" />
              </div>
              <div className="mb-8 p-4 rounded-xl bg-accent-orange/10 border border-accent-orange/20 w-fit text-accent-orange shadow-[0_0_20px_rgba(242,123,47,0.1)]">
                {feature.icon}
              </div>
              <h3 className="text-lg font-bold text-white uppercase tracking-tight mb-4">
                {feature.title}
              </h3>
              <p className="text-sm text-muted leading-relaxed font-medium">
                {feature.desc}
              </p>
            </motion.div>
          ))}
        </motion.div>
      </section>

      {/* 📊 The Workflow - Unified UX */}
      <section className="relative overflow-hidden bg-white/2 py-32 border-y border-white/5">
        <div className="app-container grid lg:grid-cols-2 gap-20 items-center">
          <motion.div {...fadeInUp} className="space-y-10">
            <div className="pill border-accent-orange/20 text-accent-orange">
              Operational Stream
            </div>
            <h2 className="text-5xl font-black text-white leading-tight tracking-tighter">
              A Seamless <br />
              <span className="text-accent-orange">Technical Pipeline.</span>
            </h2>
            <div className="space-y-8">
              {[
                "Instant Ingestion via Paste or Excel Stream",
                "Validation Engine with Real-time Analysis",
                "Single-click Markdown Serialization",
                "Zero-Login Architecture (Session Based)",
              ].map((text, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 group cursor-default"
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white/5 border border-white/10 group-hover:border-accent-orange group-hover:bg-accent-orange/10 transition-all">
                    <CheckCircle2 className="w-5 h-5 text-accent-orange" />
                  </div>
                  <span className="text-base font-bold text-white/70 group-hover:text-white transition-colors">
                    {text}
                  </span>
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="relative"
          >
            <div className="absolute -inset-4 bg-accent-orange/10 blur-[80px] rounded-full animate-float" />
            <div className="relative surface p-0! overflow-hidden border-white/10 shadow-3xl">
              <div className="bg-white/5 border-b border-white/10 p-4 flex items-center justify-between">
                <div className="flex gap-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-accent-red/40" />
                  <div className="w-2.5 h-2.5 rounded-full bg-accent-gold/40" />
                  <div className="w-2.5 h-2.5 rounded-full bg-accent-orange/40" />
                </div>
                <span className="text-[10px] font-bold uppercase tracking-widest text-muted">
                  engine.telemetry.01
                </span>
              </div>
              <div className="p-8 space-y-6">
                <div className="h-4 w-3/4 bg-white/5 rounded-full" />
                <div className="h-4 w-1/2 bg-white/5 rounded-full" />
                <div className="h-4 w-5/6 bg-white/5 rounded-full" />
                <div className="h-4 w-2/3 bg-white/5 rounded-full" />
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* 🏁 Final CTA */}
      <section className="app-container flex flex-col items-center text-center py-20 px-4 space-y-12">
        <motion.div {...fadeInUp} className="max-w-2xl space-y-6">
          <h2 className="text-5xl font-black text-white tracking-tighter uppercase">
            Ready to <span className="text-accent-orange">Convert?</span>
          </h2>
          <p className="text-muted text-lg font-medium">
            Join developers and engineers standardizing their documentation
            workflow. Instant, secure, and purely technical.
          </p>
        </motion.div>

        <motion.div {...fadeInUp} className="pt-10">
          <Link href="/app">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.98 }}
              className="btn-primary h-20 px-16 text-xl shadow-2xl shadow-accent-orange/20"
            >
              Start Your Workbench
            </motion.button>
          </Link>
          <div className="mt-10 flex items-center justify-center gap-10 text-[10px] font-bold uppercase tracking-[0.4em] text-white/60">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4" /> Scalable
            </div>
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4" /> Local
            </div>
            <div className="flex items-center gap-2">
              <Zap className="w-4 h-4" /> Fast
            </div>
          </div>
        </motion.div>
      </section>
    </div>
  );
}
