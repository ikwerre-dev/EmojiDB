"use client";

import DecorativeBorder from "@/components/DecorativeBorder";
import DiagonalMesh from "@/components/DiagonalMesh";
import CodeEditor from "@/components/CodeEditor";
import { Copy, Github, SmileIcon, User2 } from "lucide-react";

export default function Home() {
  const copyToClipboard = () => {
    navigator.clipboard.writeText('npm install @ikwerre-dev/emojidb');
  };

  return (
    <div className="flex min-h-screen lg:px-5 flex-col  bg-[#0a0a0a] font-sans">
      <DecorativeBorder side="left" />
      <DecorativeBorder side="right" />

      <div className="px-4 sm:px-8 lg:px-16 min-h-screen flex lg:py-[1rem] flex-col w-full">
        <main className="w-full ">
          <h1 className="font-sekuya py-2 md:py-0 lg:border-x-[2.3px] lg:border-dashed border-white/90 font-[900] text-center text-[70px] sm:text-[120px] md:text-[160px] lg:text-[260px]">
            <span className="flex flex-col  leading-[1.3]"  >
              <span className="text-white/80">EmojiDB</span>
            </span>
          </h1>
          <div className="grid grid-cols-1 md:grid-cols-6 border-2 border-x-white/90">
            <a
              href="https://github.com/ikwerre-dev/emojidb"
              target="_blank"
              rel="noopener noreferrer"
              className="border-b md:border-b-0 md:border-r-2 col-span-1 md:col-span-2 gap-2 flex cursor-pointer bg-white/5 hover:bg-white/50 transition-all duration-300 p-3 md:p-5"
            >
              <Github size={18} className="md:size-5" />
              <p className="text-sm md:text-base">Star this shit project on Github</p>
            </a>

            <div className="flex gap-3 md:gap-5 col-span-1 md:col-span-4 px-4 md:px-10 text-sm md:text-base font-sekuya items-center justify-center md:justify-end p-3 md:p-5">
              <p className="truncate">npm install @ikwerre-dev/emojidb</p>
              <button onClick={copyToClipboard} className="hover:opacity-70 transition-opacity">
                <Copy size={16} className="md:size-[18px] flex-shrink-0" />
              </button>
            </div>
          </div>
        </main>



        <div className="grid grid-cols-1 md:grid-cols-6 border-x-2 border-x-white/90">
          <div className="hidden md:block col-span-1 border-b-2 border-b-white/90 border-r-2 border-r-white/90">
          </div>
          <div className="col-span-1 md:col-span-3 border-b-2 py-6 md:py-5 border-b-white/90 md:border-r-2 md:border-r-white/90 flex flex-col justify-center px-6 md:px-8">
            <h2 className="font-sekuya text-3xl md:text-4xl lg:text-5xl font-bold text-white/90 mb-3 md:mb-4">
              Database, but encrypted with emojis
            </h2>
            <p className="text-white/70 text-sm md:text-base leading-relaxed">
              A lightweight, secure database that encrypts your data and encodes it into emojis.
              Fast queries, simple API, and built-in encryption make it perfect for modern applications.
            </p>
          </div>
          <div className="hidden md:flex col-span-1 flex-col items-center justify-center text-center border-b-2 border-b-white/90 border-r-2 border-r-white/90">
          </div>
          <div className="hidden md:block col-span-1 border-b-2 border-b-white/90">
            <div className="h-full border-t-[12px] border-r-[12px] border-t-[#4d4d4d] border-r-[#4d4d4d]">
              <DiagonalMesh />
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-6 border-x-2 pt-[.5px] border-x-white/90">
          <div className="hidden md:block col-span-1 h-[30rem] border-b-2 border-b-white/90 border-r-2 border-r-white/90">
            <div className="h-full border-t-[12px] border-r-[12px] border-t-[#4d4d4d] border-r-[#4d4d4d]">
              <DiagonalMesh />
            </div>
          </div>
          <div className="col-span-1 md:col-span-3 h-[20rem] md:h-[30rem] border-b-2 border-b-white/90 md:border-r-2 md:border-r-white/90 relative">
            <CodeEditor />
          </div>
          <div className="col-span-1 md:col-span-2 min-h-[20rem] md:h-[30rem] border-b-2 border-b-white/90 flex flex-col px-6 md:px-8 py-6">
            <h3 className="font-sekuya text-2xl md:text-3xl font-bold text-white/90 mb-4 md:mb-6">How it Works</h3>
            <div className="space-y-4">
              <div className="flex gap-3">
                <div className="flex-shrink-0 items-center text-sm font-bold">
                  -
                </div>
                <div>
                  <h4 className="font-semibold text-white/90 text-base mb-0.5">Insert Data</h4>
                  <p className="text-sm text-white/60">Your data is received via the simple API</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 items-center text-base font-bold">
                  -
                </div>
                <div>
                  <h4 className="font-semibold text-white/90 text-base mb-0.5">Encrypt in Memory</h4>
                  <p className="text-sm text-white/60">Data is encrypted using AES-256 with your secret key</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 items-center text-base font-bold">
                  -
                </div>
                <div>
                  <h4 className="font-semibold text-white/90 text-base mb-0.5">Convert to Bytes</h4>
                  <p className="text-sm text-white/60">Encrypted data is converted to byte sequences</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 items-center text-base font-bold">
                  -
                </div>
                <div>
                  <h4 className="font-semibold text-white/90 text-base mb-0.5">Encode to Emoji Pairs</h4>
                  <p className="text-sm text-white/60">Each byte pair is mapped to a unique emoji combination</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 items-center text-base font-bold">
                  -
                </div>
                <div>
                  <h4 className="font-semibold text-white/90 text-base mb-0.5">Store & Query</h4>
                  <p className="text-sm text-white/60">Emoji-encoded data is stored for fast retrieval</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-6 border-x-2 border-x-white/90">
          <div className="hidden md:block col-span-1 border-b-2 border-b-white/90 border-r-2 border-r-white/90"></div>

          <div className="col-span-1 md:col-span-2 border-b-2 border-b-white/90 border-r-2 border-r-white/90 bg-[#1e1e1e] p-6">
            <h3 className="font-sekuya text-xl font-bold text-white/90 mb-4">Original Data</h3>
            <pre className="text-xs md:text-sm text-green-400 font-mono overflow-x-auto">
              {`[
  {
    "id": "1",
    "email": "honour@example.com",
    "name": "Robinson Honour"
  },
  {
    "id": "2",
    "email": "john@example.com",
    "name": "John Doe"
  }
]`}
            </pre>
          </div>

          <div className="col-span-1 md:col-span-2 border-b-2 border-b-white/90 border-r-2 border-r-white/90 bg-[#0a0a0a] p-6">
            <h3 className="font-sekuya text-xl font-bold text-white/90 mb-4">Emoji Encoded</h3>
            <div className="grid grid-cols-10 gap-1">
              {["😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "😍", "😘", "🥰", "😗",
                "😙", "😚", "☺️", "🙂", "🤗", "🤩", "🤔", "🤨", "😐", "😑", "😶", "🙄", "😏", "😣", "😥", "😮",
                "🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓", "😔", "😕", "🙃", "🤑",
                "😲", "☹️", "🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩", "🤯", "😬", "😰",
                "😱", "🥵", "🥶", "😳", "🤪", "😵"].map((emoji, i) => (
                  <div key={i} className="text-2xl text-center hover:scale-110 transition-transform">
                    {emoji}
                  </div>
                ))}
            </div>
          </div>

          <div className="hidden md:block col-span-1 border-b-2 border-b-white/90"></div>
        </div>

        <div className="relative border-x-2 border-x-white/90">
          <div className="absolute inset-0">
            <svg width="100%" height="100%" aria-hidden="true">
              <defs>
                <pattern viewBox="0 0 10 10" width="10" height="10" patternUnits="userSpaceOnUse" id="_r12R_6_">
                  <circle cx="5" cy="5" r="1" fill="currentColor" className="fill-white/30"></circle>
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#_r12R_6_)"></rect>
            </svg>
          </div>

          <div className="relative z-10 p-4 md:p-8">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-0 bg-[#0a0a0a]">
              <div className="col-span-1 flex flex-col items-center justify-center py-8 md:py-12">
                <p className="font-sekuya text-3xl md:text-4xl font-bold text-white/90 mb-2">10,994</p>
                <p className="text-white/60 text-sm">Lines of Code</p>
              </div>

              <div className="col-span-1 flex flex-col items-center justify-center py-8 md:py-12">
                <p className="font-sekuya text-3xl md:text-4xl font-bold text-white/90 mb-2">40</p>
                <p className="text-white/60 text-sm">Code Files</p>
              </div>

              <div className="col-span-1 flex flex-col items-center justify-center py-8 md:py-12">
                <p className="font-sekuya text-3xl md:text-4xl font-bold text-white/90 mb-2">48h</p>
                <p className="text-white/60 text-sm">Build Time</p>
              </div>
            </div>
          </div>
        </div>


        <div className="grid grid-cols-1 md:grid-cols-6 border-x-2 border-x-white/90">
          <div className="col-span-1 md:col-span-2 border-y-2 border-y-white/90 md:border-r-2 md:border-r-white/90 flex flex-col px-6 md:px-8 py-6">
            <h3 className="font-sekuya text-2xl font-bold text-white/90 mb-4">Why did I Build This</h3>
            <p className="text-white/70 text-sm leading-relaxed">
              Honestly? I was bored and thought it would be cool to encrypt data and encode it as emojis.
              Turns out it actually works pretty well. The full stuff was built under 48 hours.
            </p>
          </div>

          <div className="hidden md:block col-span-1 border-y-2 border-y-white/90 border-r-2 border-r-white/90">
            <div className="h-full border-t-[12px] border-r-[12px] border-t-[#4d4d4d] border-r-[#4d4d4d]">
              <DiagonalMesh />
            </div>
          </div>

          <div className="col-span-1 border-y-2 border-y-white/90 md:border-r-2 md:border-r-white/90 flex flex-col px-6 py-6">
            <h3 className="font-sekuya text-lg md:text-xl font-bold text-white/90 mb-4">Tech Stack</h3>
            <p className="text-white/70 text-sm">
              Go, TypeScript, Node.js
            </p>
          </div>

          <div className="col-span-1 border-y-2 border-y-white/90 md:border-r-2 md:border-r-white/90 flex flex-col px-6 py-6">
            <h3 className="font-sekuya text-lg md:text-xl font-bold text-white/90 mb-4">Links</h3>
            <div className="space-y-3">
              <a
                href="https://github.com/ikwerre-dev/emojidb"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-white/80 hover:text-white transition-colors"
              >
                <Github size={16} />
                <span className="text-xs">GitHub</span>
              </a>
              <a
                href="https://robinsonhonour.me"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-white/80 hover:text-white transition-colors"
              >
                <User2 size={16} />
                <span className="text-xs">Portfolio</span>
              </a>
            </div>
          </div>

          <div className="col-span-1 border-y-2 border-y-white/90 flex flex-col px-6 py-6">
            <p className="text-white/60 text-xs mb-2">© 2025 EmojiDB</p>
            <p className="text-white/50 text-xs">Built by Robinson Honour</p>
          </div>
        </div>

      </div>
    </div>
  );
}
