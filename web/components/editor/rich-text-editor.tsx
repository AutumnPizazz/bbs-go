"use client"

import * as React from "react"
import { createPortal } from "react-dom"
import { Extension, Node, mergeAttributes, type CommandProps, type Editor, type Range } from "@tiptap/core"
import { EditorContent, NodeViewWrapper, ReactNodeViewRenderer, ReactRenderer, type ReactNodeViewProps, useEditor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Link from "@tiptap/extension-link"
import Placeholder from "@tiptap/extension-placeholder"
import Underline from "@tiptap/extension-underline"
import TextAlign from "@tiptap/extension-text-align"
import { TextStyle } from "@tiptap/extension-text-style"
import BackgroundColor from "@tiptap/extension-text-style/background-color"
import Color from "@tiptap/extension-color"
import TaskList from "@tiptap/extension-task-list"
import TaskItem from "@tiptap/extension-task-item"
import Typography from "@tiptap/extension-typography"
import HorizontalRule from "@tiptap/extension-horizontal-rule"
import Suggestion, { exitSuggestion, type SuggestionKeyDownProps, type SuggestionProps } from "@tiptap/suggestion"
import { Table } from "@tiptap/extension-table"
import { TableRow } from "@tiptap/extension-table-row"
import { TableCell } from "@tiptap/extension-table-cell"
import { TableHeader } from "@tiptap/extension-table-header"
import { Highlight } from "@tiptap/extension-highlight"
import { FontSize } from "@tiptap/extension-font-size"
import { Superscript } from "@tiptap/extension-superscript"
import { Subscript } from "@tiptap/extension-subscript"
import { Mention } from "@tiptap/extension-mention"
import { CodeBlockLowlight } from "@tiptap/extension-code-block-lowlight"
import { common, createLowlight } from "lowlight"
import { PluginKey } from "@tiptap/pm/state"
import {
  AlignCenter,
  AlignLeft,
  AlignRight,
  ArrowLeft,
  ArrowRight,
  Bold,
  Check,
  Code,
  Code2,
  Eraser,
  Heading1,
  Heading2,
  Heading3,
  Highlighter,
  ImageIcon,
  Italic,
  LinkIcon,
  List,
  ListOrdered,
  ListTodo,
  Maximize,
  Minimize,
  MinusSquare,
  Paintbrush,
  Palette,
  Pilcrow,
  Plus,
  Quote,
  Redo2,
  Smile,
  Strikethrough,
  Subscript as SubscriptIcon,
  Superscript as SuperscriptIcon,
  Table as TableIcon,
  Table2,
  Trash2,
  Underline as UnderlineIcon,
  Undo2,
} from "lucide-react"

import { uploadEditorImage } from "@/components/editor/upload"
import { searchUsers } from "@/lib/api/users"
import type { SearchUser } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { useToastActions } from "@/lib/toast"
import { cn } from "@/lib/utils"

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    resizableImage: {
      setResizableImage: (options: { src: string; alt?: string; title?: string; width?: number; height?: number }) => ReturnType
    }
  }
}

const TEXT_COLOR_PALETTE = [
  "#c00000",
  "#ff0000",
  "#ffc000",
  "#ffff00",
  "#a5d610",
  "#00b050",
  "#00b0f0",
  "#0070c0",
  "#002060",
  "#7030a0",
  "#ffffff",
  "#000000",
  "#eeeeee",
  "#525252",
  "#1890ff",
  "#ff7875",
  "#52c41a",
  "#fa8c16",
  "#722ed1",
  "#eb2f96",
]

const BACKGROUND_COLOR_PALETTE = [
  "#FFCCCC",
  "#FFE6CC",
  "#FFFFCC",
  "#CCFFCC",
  "#CCFFFF",
  "#CCE5FF",
  "#E5CCFF",
  "#FFCCFF",
  "#F2F2F2",
  "#E6E6E6",
  "#FFD6CC",
  "#E5FFCC",
  "#CCFFE5",
  "#D6FFFF",
  "#FFE0F2",
  "#FFF0F0",
  "#FFF9E6",
  "#F0FFF0",
  "#F0FFFF",
  "#F5FFFA",
]

const lowlight = createLowlight(common)

const FONT_SIZES = [
  { label: "12px", value: "12px" },
  { label: "14px", value: "14px" },
  { label: "16px", value: "16px" },
  { label: "18px", value: "18px" },
  { label: "20px", value: "20px" },
  { label: "24px", value: "24px" },
  { label: "28px", value: "28px" },
  { label: "32px", value: "32px" },
]

const EMOJI_LIST = [
  { emoji: "😀", keywords: ["smile", "happy", "grin"] },
  { emoji: "😂", keywords: ["joy", "laugh", "tear"] },
  { emoji: "🤣", keywords: ["rofl", "rolling"] },
  { emoji: "😍", keywords: ["heart", "eyes", "love"] },
  { emoji: "😎", keywords: ["cool", "sunglasses"] },
  { emoji: "🥳", keywords: ["party", "celebrate"] },
  { emoji: "😢", keywords: ["cry", "sad", "tear"] },
  { emoji: "😡", keywords: ["angry", "mad", "pout"] },
  { emoji: "👍", keywords: ["thumbsup", "like", "+1"] },
  { emoji: "👎", keywords: ["thumbsdown", "dislike"] },
  { emoji: "👏", keywords: ["clap", "applause"] },
  { emoji: "🙏", keywords: ["pray", "thanks", "please"] },
  { emoji: "💪", keywords: ["strong", "muscle"] },
  { emoji: "🔥", keywords: ["fire", "hot", "lit"] },
  { emoji: "💯", keywords: ["100", "perfect", "hundred"] },
  { emoji: "🎉", keywords: ["party", "celebrate", "tada"] },
  { emoji: "🎊", keywords: ["confetti", "celebrate"] },
  { emoji: "❤️", keywords: ["heart", "love"] },
  { emoji: "💔", keywords: ["heartbreak", "broken"] },
  { emoji: "⭐", keywords: ["star", "favorite"] },
  { emoji: "✅", keywords: ["check", "done", "ok"] },
  { emoji: "❌", keywords: ["cross", "wrong", "no"] },
  { emoji: "❓", keywords: ["question", "what"] },
  { emoji: "❗", keywords: ["exclamation", "warn"] },
  { emoji: "💡", keywords: ["idea", "tip", "lightbulb"] },
  { emoji: "📌", keywords: ["pin", "pushpin"] },
  { emoji: "📎", keywords: ["paperclip", "attach"] },
  { emoji: "🔗", keywords: ["link", "chain"] },
  { emoji: "🚀", keywords: ["rocket", "launch"] },
  { emoji: "🐛", keywords: ["bug", "insect"] },
  { emoji: "🧠", keywords: ["brain", "mind"] },
  { emoji: "🤖", keywords: ["robot", "ai"] },
  { emoji: "✨", keywords: ["sparkles", "magic", "shine"] },
  { emoji: "💻", keywords: ["computer", "laptop"] },
  { emoji: "📱", keywords: ["phone", "mobile"] },
  { emoji: "🖥️", keywords: ["desktop", "monitor"] },
  { emoji: "⌨️", keywords: ["keyboard"] },
  { emoji: "🐧", keywords: ["penguin", "linux"] },
  { emoji: "☕", keywords: ["coffee", "java"] },
  { emoji: "🍺", keywords: ["beer", "drink"] },
  { emoji: "🎯", keywords: ["target", "goal"] },
  { emoji: "🏆", keywords: ["trophy", "win"] },
  { emoji: "🔒", keywords: ["lock", "secure"] },
  { emoji: "🔑", keywords: ["key", "password"] },
  { emoji: "📝", keywords: ["memo", "write", "note"] },
  { emoji: "✏️", keywords: ["pencil", "edit"] },
  { emoji: "🗑️", keywords: ["trash", "delete"] },
  { emoji: "📦", keywords: ["package", "box"] },
  { emoji: "🧩", keywords: ["puzzle", "piece"] },
  { emoji: "🔍", keywords: ["search", "magnify"] },
  { emoji: "🌐", keywords: ["web", "globe", "internet"] },
]

type Translate = (key: string) => string

type RichTextEditorLabels = {
  placeholder: string
  toolbar: {
    undo: string
    redo: string
    bold: string
    underline: string
    italic: string
    strike: string
    heading1: string
    heading2: string
    quote: string
    bulletList: string
    orderedList: string
    taskList: string
    alignLeft: string
    alignCenter: string
    alignRight: string
    textColor: string
    backgroundColor: string
    clearColor: string
    inlineCode: string
    codeBlock: string
    link: string
    image: string
    horizontalRule: string
    table: string
    highlight: string
    fontSize: string
    superscript: string
    subscript: string
    clearFormat: string
    addRowBefore: string
    addRowAfter: string
    addColBefore: string
    addColAfter: string
    deleteRow: string
    deleteCol: string
    deleteTable: string
    fullscreen: string
    exitFullscreen: string
    uploading: string
  }
  linkDialog: {
    textLabel: string
    urlLabel: string
    textPlaceholder: string
    urlPlaceholder: string
    confirm: string
    remove: string
    cancel: string
  }
  slash: {
    label: string
    hintContinueTyping: string
    noMatchingCommand: string
    paragraph: { title: string; description: string }
    heading1: { title: string; description: string }
    heading2: { title: string; description: string }
    heading3: { title: string; description: string }
    bulletList: { title: string; description: string }
    orderedList: { title: string; description: string }
    taskList: { title: string; description: string }
    quote: { title: string; description: string }
    codeBlock: { title: string; description: string }
    horizontalRule: { title: string; description: string }
    calloutInfo: { title: string; description: string }
    calloutWarning: { title: string; description: string }
    calloutTip: { title: string; description: string }
    calloutSuccess: { title: string; description: string }
  }
}

function createEditorLabels(t: Translate): RichTextEditorLabels {
  const key = (path: string) => `component.richTextEditor.${path}`

  return {
    placeholder: t(key("placeholder")),
    toolbar: {
      undo: t(key("toolbar.undo")),
      redo: t(key("toolbar.redo")),
      bold: t(key("toolbar.bold")),
      underline: t(key("toolbar.underline")),
      italic: t(key("toolbar.italic")),
      strike: t(key("toolbar.strike")),
      heading1: t(key("toolbar.heading1")),
      heading2: t(key("toolbar.heading2")),
      quote: t(key("toolbar.quote")),
      bulletList: t(key("toolbar.bulletList")),
      orderedList: t(key("toolbar.orderedList")),
      taskList: t(key("toolbar.taskList")),
      alignLeft: t(key("toolbar.alignLeft")),
      alignCenter: t(key("toolbar.alignCenter")),
      alignRight: t(key("toolbar.alignRight")),
      textColor: t(key("toolbar.textColor")),
      backgroundColor: t(key("toolbar.backgroundColor")),
      clearColor: t(key("toolbar.clearColor")),
      inlineCode: t(key("toolbar.inlineCode")),
      codeBlock: t(key("toolbar.codeBlock")),
      link: t(key("toolbar.link")),
      image: t(key("toolbar.image")),
      horizontalRule: t(key("toolbar.horizontalRule")),
      table: t(key("toolbar.table")),
      highlight: t(key("toolbar.highlight")),
      fontSize: t(key("toolbar.fontSize")),
      superscript: t(key("toolbar.superscript")),
      subscript: t(key("toolbar.subscript")),
      clearFormat: t(key("toolbar.clearFormat")),
      addRowBefore: t(key("toolbar.addRowBefore")),
      addRowAfter: t(key("toolbar.addRowAfter")),
      addColBefore: t(key("toolbar.addColBefore")),
      addColAfter: t(key("toolbar.addColAfter")),
      deleteRow: t(key("toolbar.deleteRow")),
      deleteCol: t(key("toolbar.deleteCol")),
      deleteTable: t(key("toolbar.deleteTable")),
      fullscreen: t(key("toolbar.fullscreen")),
      exitFullscreen: t(key("toolbar.exitFullscreen")),
      uploading: t(key("toolbar.uploading")),
    },
    linkDialog: {
      textLabel: t(key("linkDialog.textLabel")),
      urlLabel: t(key("linkDialog.urlLabel")),
      textPlaceholder: t(key("linkDialog.textPlaceholder")),
      urlPlaceholder: t(key("linkDialog.urlPlaceholder")),
      confirm: t(key("linkDialog.confirm")),
      remove: t(key("linkDialog.remove")),
      cancel: t(key("linkDialog.cancel")),
    },
    slash: {
      label: t("common.accessibility.slashCommands"),
      hintContinueTyping: t(key("slash.hintContinueTyping")),
      noMatchingCommand: t(key("slash.noMatchingCommand")),
      paragraph: {
        title: t(key("slash.paragraph.title")),
        description: t(key("slash.paragraph.description")),
      },
      heading1: {
        title: t(key("slash.heading1.title")),
        description: t(key("slash.heading1.description")),
      },
      heading2: {
        title: t(key("slash.heading2.title")),
        description: t(key("slash.heading2.description")),
      },
      heading3: {
        title: t(key("slash.heading3.title")),
        description: t(key("slash.heading3.description")),
      },
      bulletList: {
        title: t(key("slash.bulletList.title")),
        description: t(key("slash.bulletList.description")),
      },
      orderedList: {
        title: t(key("slash.orderedList.title")),
        description: t(key("slash.orderedList.description")),
      },
      taskList: {
        title: t(key("slash.taskList.title")),
        description: t(key("slash.taskList.description")),
      },
      quote: {
        title: t(key("slash.quote.title")),
        description: t(key("slash.quote.description")),
      },
      codeBlock: {
        title: t(key("slash.codeBlock.title")),
        description: t(key("slash.codeBlock.description")),
      },
      horizontalRule: {
        title: t(key("slash.horizontalRule.title")),
        description: t(key("slash.horizontalRule.description")),
      },
      calloutInfo: {
        title: t(key("slash.calloutInfo.title")),
        description: t(key("slash.calloutInfo.description")),
      },
      calloutWarning: {
        title: t(key("slash.calloutWarning.title")),
        description: t(key("slash.calloutWarning.description")),
      },
      calloutTip: {
        title: t(key("slash.calloutTip.title")),
        description: t(key("slash.calloutTip.description")),
      },
      calloutSuccess: {
        title: t(key("slash.calloutSuccess.title")),
        description: t(key("slash.calloutSuccess.description")),
      },
    },
  }
}

const RESIZE_HANDLES = ["nw", "n", "ne", "e", "se", "s", "sw", "w"]

function ResizableImageView({
  node,
  selected,
  editor,
  updateAttributes,
  getPos,
}: ReactNodeViewProps) {
  const attrs = node.attrs as { src: string; alt?: string; title?: string; width?: number; height?: number }
  const imageRef = React.useRef<HTMLImageElement>(null)
  const [imageLoaded, setImageLoaded] = React.useState(false)
  const [size, setSize] = React.useState<{ width?: number; height?: number }>({
    width: attrs.width,
    height: attrs.height,
  })
  const aspectRatioRef = React.useRef(1)
  const resizeRef = React.useRef<{
    handle: string
    startX: number
    startWidth: number
    startHeight: number
  } | null>(null)

  function onImageLoad() {
    const image = imageRef.current
    if (!image) return
    setImageLoaded(true)
    aspectRatioRef.current = image.naturalWidth / image.naturalHeight || 1
    if (!size.width && !size.height) {
      const maxWidth = 600
      const width = image.naturalWidth > maxWidth ? maxWidth : image.naturalWidth
      const height = Math.round(width / aspectRatioRef.current)
      setSize({ width, height })
      updateAttributes({ width, height })
    }
  }

  function startResize(event: React.MouseEvent, handle: string) {
    event.preventDefault()
    event.stopPropagation()
    const image = imageRef.current
    if (!image) return
    resizeRef.current = {
      handle,
      startX: event.clientX,
      startWidth: size.width || image.offsetWidth,
      startHeight: size.height || image.offsetHeight,
    }
    editor.view.dom.classList.add("resizing-image")
  }

  React.useEffect(() => {
    function onMouseMove(event: MouseEvent) {
      const current = resizeRef.current
      if (!current) return
      event.preventDefault()
      const deltaX = event.clientX - current.startX
      let width = current.startWidth
      if (["se", "e", "s", "ne"].includes(current.handle)) {
        width = current.startWidth + deltaX
      } else {
        width = current.startWidth - deltaX
      }
      width = Math.max(50, Math.min(800, width))
      const height = Math.round(width / aspectRatioRef.current)
      setSize({ width, height })
    }

    function onMouseUp() {
      const current = resizeRef.current
      if (!current) return
      resizeRef.current = null
      editor.view.dom.classList.remove("resizing-image")
      updateAttributes({ width: size.width, height: size.height })
    }

    document.addEventListener("mousemove", onMouseMove)
    document.addEventListener("mouseup", onMouseUp)
    return () => {
      document.removeEventListener("mousemove", onMouseMove)
      document.removeEventListener("mouseup", onMouseUp)
    }
  }, [editor, size.height, size.width, updateAttributes])

  function selectImage() {
    const pos = getPos()
    if (typeof pos === "number") {
      editor.commands.setNodeSelection(pos)
    }
  }

  return (
    <NodeViewWrapper className={cn("resizable-image-wrapper", selected && "is-selected")}>      <img
        ref={imageRef}
        src={attrs.src}
        alt={attrs.alt || ""}
        title={attrs.title || ""}
        width={size.width}
        height={size.height}
        className="editor-image resizable"
        onLoad={onImageLoad}
        onClick={selectImage}
      />
      {selected && imageLoaded ? (
        <>
          <div className="selection-border">
            <div className="border-line border-top" />
            <div className="border-line border-right" />
            <div className="border-line border-bottom" />
            <div className="border-line border-left" />
          </div>
          {RESIZE_HANDLES.map((handle) => (
            <div key={handle} className={`resize-handle resize-handle-${handle}`} onMouseDown={(event) => startResize(event, handle)} />
          ))}
        </>
      ) : null}
    </NodeViewWrapper>
  )
}

const ResizableImage = Node.create({
  name: "resizableImage",
  group: "block",
  draggable: true,

  addAttributes() {
    return {
      src: { default: null },
      alt: { default: null },
      title: { default: null },
      width: {
        default: null,
        parseHTML: (element: HTMLElement) => {
          const width = element.getAttribute("width")
          return width ? Number.parseInt(width, 10) : null
        },
      },
      height: {
        default: null,
        parseHTML: (element: HTMLElement) => {
          const height = element.getAttribute("height")
          return height ? Number.parseInt(height, 10) : null
        },
      },
    }
  },

  parseHTML() {
    return [{ tag: "img[src]" }]
  },

  renderHTML({ HTMLAttributes }: { HTMLAttributes: Record<string, unknown> }) {
    return ["img", mergeAttributes(HTMLAttributes)]
  },

  addCommands() {
    return {
      setResizableImage:
        (options: { src: string; alt?: string; title?: string; width?: number; height?: number }) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: this.name,
            attrs: options,
          }),
    }
  },

  addNodeView() {
    return ReactNodeViewRenderer(ResizableImageView)
  },
})

const CALLOUT_CONFIG = {
  info: {
    icon: "ℹ️",
    class: "callout-info",
    zhLabel: "信息",
    enLabel: "Info",
  },
  warning: {
    icon: "⚠️",
    class: "callout-warning",
    zhLabel: "警告",
    enLabel: "Warning",
  },
  tip: {
    icon: "💡",
    class: "callout-tip",
    zhLabel: "提示",
    enLabel: "Tip",
  },
  success: {
    icon: "✅",
    class: "callout-success",
    zhLabel: "成功",
    enLabel: "Success",
  },
} as const

type CalloutType = keyof typeof CALLOUT_CONFIG

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    callout: {
      setCallout: (options: { type?: CalloutType }) => ReturnType
    }
  }
}

const Callout = Node.create({
  name: "callout",
  group: "block",
  content: "block+",
  defining: true,

  addAttributes() {
    return {
      calloutType: {
        default: "info",
        parseHTML: (element: HTMLElement) => {
          for (const key of Object.keys(CALLOUT_CONFIG)) {
            if (element.classList.contains(CALLOUT_CONFIG[key as CalloutType].class)) {
              return key
            }
          }
          return "info"
        },
      },
    }
  },

  parseHTML() {
    return [
      {
        tag: "div.callout",
        getAttrs: (element: string | HTMLElement) => {
          if (typeof element === "string") return { calloutType: "info" }
          for (const key of Object.keys(CALLOUT_CONFIG)) {
            if (element.classList.contains(CALLOUT_CONFIG[key as CalloutType].class)) {
              return { calloutType: key }
            }
          }
          return { calloutType: "info" }
        },
      },
    ]
  },

  renderHTML({ HTMLAttributes, node }: { HTMLAttributes: Record<string, unknown>; node: { attrs: Record<string, unknown> } }) {
    const calloutType = (node.attrs.calloutType as CalloutType) || "info"
    const config = CALLOUT_CONFIG[calloutType] || CALLOUT_CONFIG.info
    return [
      "div",
      mergeAttributes(HTMLAttributes, {
        class: `callout ${config.class}`,
        "data-type": calloutType,
      }),
      ["div", { class: "callout-icon" }, config.icon],
      ["div", { class: "callout-content" }, 0],
    ]
  },

  addCommands() {
    return {
      setCallout:
        (options?: { type?: CalloutType }) =>
        ({ commands }: CommandProps) =>
          commands.wrapIn(this.name, { calloutType: options?.type || "info" }),
    }
  },
})

type SlashCommandItem = {
  title: string
  description: string
  aliases: string[]
  icon: React.ReactNode
  command: ({ editor, range }: { editor: Editor; range: Range }) => void
}

type SlashCommandMenuProps = SuggestionProps<SlashCommandItem, SlashCommandItem> & {
  labels: RichTextEditorLabels
}

type SlashCommandMenuHandle = {
  onKeyDown: (event: KeyboardEvent) => boolean
}

function slashItems(labels: RichTextEditorLabels, locale: string): SlashCommandItem[] {
  const zh = locale === "zh-CN"
  const aliases = zh
    ? {
        paragraph: ["p", "text", "paragraph", "正文", "段落"],
        heading1: ["h1", "一级标题", "heading1"],
        heading2: ["h2", "二级标题", "heading2"],
        heading3: ["h3", "三级标题", "heading3"],
        bulletList: ["ul", "bullet", "list", "无序列表"],
        orderedList: ["ol", "ordered", "numbered", "有序列表"],
        taskList: ["task", "todo", "checklist", "任务列表", "待办"],
        quote: ["quote", "blockquote", "引用", "引用文本"],
        codeBlock: ["code", "codeblock", "代码", "代码块"],
        horizontalRule: ["hr", "line", "divider", "分割线"],
      }
    : {
        paragraph: ["p", "text", "paragraph"],
        heading1: ["h1", "heading1"],
        heading2: ["h2", "heading2"],
        heading3: ["h3", "heading3"],
        bulletList: ["ul", "bullet", "list"],
        orderedList: ["ol", "ordered", "numbered"],
        taskList: ["task", "todo", "checklist"],
        quote: ["quote", "blockquote"],
        codeBlock: ["code", "codeblock"],
        horizontalRule: ["hr", "line", "divider"],
      }

  return [
    {
      title: labels.slash.paragraph.title,
      description: labels.slash.paragraph.description,
      aliases: aliases.paragraph,
      icon: <Pilcrow size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setParagraph().run(),
    },
    {
      title: labels.slash.heading1.title,
      description: labels.slash.heading1.description,
      aliases: aliases.heading1,
      icon: <Heading1 size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode("heading", { level: 1 }).run(),
    },
    {
      title: labels.slash.heading2.title,
      description: labels.slash.heading2.description,
      aliases: aliases.heading2,
      icon: <Heading2 size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode("heading", { level: 2 }).run(),
    },
    {
      title: labels.slash.heading3.title,
      description: labels.slash.heading3.description,
      aliases: aliases.heading3,
      icon: <Heading3 size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode("heading", { level: 3 }).run(),
    },
    {
      title: labels.slash.bulletList.title,
      description: labels.slash.bulletList.description,
      aliases: aliases.bulletList,
      icon: <List size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBulletList().run(),
    },
    {
      title: labels.slash.orderedList.title,
      description: labels.slash.orderedList.description,
      aliases: aliases.orderedList,
      icon: <ListOrdered size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleOrderedList().run(),
    },
    {
      title: labels.slash.taskList.title,
      description: labels.slash.taskList.description,
      aliases: aliases.taskList,
      icon: <ListTodo size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleTaskList().run(),
    },
    {
      title: labels.slash.quote.title,
      description: labels.slash.quote.description,
      aliases: aliases.quote,
      icon: <Quote size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBlockquote().run(),
    },
    {
      title: labels.slash.codeBlock.title,
      description: labels.slash.codeBlock.description,
      aliases: aliases.codeBlock,
      icon: <Code2 size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleCodeBlock().run(),
    },
    {
      title: labels.slash.horizontalRule.title,
      description: labels.slash.horizontalRule.description,
      aliases: aliases.horizontalRule,
      icon: <MinusSquare size={18} />,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setHorizontalRule().run(),
    },
    {
      title: labels.slash.calloutInfo.title,
      description: labels.slash.calloutInfo.description,
      aliases: zh ? ["info", "callout", "信息", "callout-info"] : ["info", "callout", "callout-info"],
      icon: <span style={{ fontSize: "18px" }}>ℹ️</span>,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setCallout({ type: "info" }).run(),
    },
    {
      title: labels.slash.calloutWarning.title,
      description: labels.slash.calloutWarning.description,
      aliases: zh ? ["warning", "warn", "警告", "callout-warning"] : ["warning", "warn", "callout-warning"],
      icon: <span style={{ fontSize: "18px" }}>⚠️</span>,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setCallout({ type: "warning" }).run(),
    },
    {
      title: labels.slash.calloutTip.title,
      description: labels.slash.calloutTip.description,
      aliases: zh ? ["tip", "hint", "提示", "callout-tip"] : ["tip", "hint", "callout-tip"],
      icon: <span style={{ fontSize: "18px" }}>💡</span>,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setCallout({ type: "tip" }).run(),
    },
    {
      title: labels.slash.calloutSuccess.title,
      description: labels.slash.calloutSuccess.description,
      aliases: zh ? ["success", "ok", "成功", "callout-success"] : ["success", "ok", "callout-success"],
      icon: <span style={{ fontSize: "18px" }}>✅</span>,
      command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setCallout({ type: "success" }).run(),
    },
  ]
}

function updateSlashMenuPosition(element: HTMLElement, clientRect?: (() => DOMRect | null) | null) {
  const rect = clientRect?.()
  if (!rect) {
    return
  }
  const gap = 8
  const offset = 6
  const popupWidth = element.offsetWidth || 320
  const popupHeight = element.offsetHeight || 360
  let left = rect.left
  let top = rect.bottom + offset

  if (left + popupWidth + gap > window.innerWidth) {
    left = window.innerWidth - popupWidth - gap
  }
  if (left < gap) {
    left = gap
  }
  if (top + popupHeight + gap > window.innerHeight) {
    top = rect.top - popupHeight - offset
  }
  if (top < gap) {
    top = gap
  }

  element.style.left = `${left}px`
  element.style.top = `${top}px`
}

const SlashCommandMenu = React.forwardRef<SlashCommandMenuHandle, SlashCommandMenuProps>(function SlashCommandMenu({ items, query, command, labels }, ref) {
  const [selectedIndex, setSelectedIndex] = React.useState(0)
  const itemRefs = React.useRef<Array<HTMLButtonElement | null>>([])

  React.useEffect(() => {
    setSelectedIndex(0)
  }, [items, query])

  React.useEffect(() => {
    setSelectedIndex((index) => {
      if (!items.length) {
        return 0
      }
      return Math.min(index, items.length - 1)
    })
  }, [items])

  React.useEffect(() => {
    itemRefs.current[selectedIndex]?.scrollIntoView({ block: "nearest" })
  }, [selectedIndex])

  React.useImperativeHandle(
    ref,
    () => ({
      onKeyDown(event) {
        if (!items.length) {
          return false
        }
        if (event.key === "ArrowUp") {
          event.preventDefault()
          setSelectedIndex((index) => (index - 1 + items.length) % items.length)
          return true
        }
        if (event.key === "ArrowDown" || event.key === "Tab") {
          event.preventDefault()
          setSelectedIndex((index) => {
            if (event.shiftKey) {
              return (index - 1 + items.length) % items.length
            }
            return (index + 1) % items.length
          })
          return true
        }
        if (event.key === "Enter") {
          event.preventDefault()
          command(items[selectedIndex])
          return true
        }
        const shortcut = Number.parseInt(event.key, 10)
        if (Number.isFinite(shortcut) && shortcut >= 1 && shortcut <= 9 && items[shortcut - 1]) {
          event.preventDefault()
          command(items[shortcut - 1])
          return true
        }
        return false
      },
    }),
    [command, items, selectedIndex]
  )

  return (
    <div className="slash-commands" role="listbox" aria-label={labels.slash.label}>
      <div className="search-hint">{items.length ? labels.slash.hintContinueTyping : labels.slash.noMatchingCommand}</div>
      {items.length ? (
        <div className="slash-items-container">
          {items.map((item, index) => (
            <button
              key={`${item.title}-${item.aliases[0] || index}`}
              type="button"
              role="option"
              aria-selected={index === selectedIndex}
              ref={(element) => {
                itemRefs.current[index] = element
              }}
              className={cn("slash-item", index === selectedIndex && "is-selected")}
              onMouseEnter={() => setSelectedIndex(index)}
              onMouseDown={(event) => {
                event.preventDefault()
                command(item)
              }}
            >
              <span className="item-icon">{item.icon}</span>
              <span className="item-content">
                <span className="item-title">
                  {item.title}
                  {item.aliases[0] ? <span className="item-aliases">/{item.aliases[0]}</span> : null}
                </span>
                <span className="item-description">{item.description}</span>
              </span>
            </button>
          ))}
        </div>
      ) : null}
    </div>
  )
})

// ---- MentionList component ----
type MentionItem = SearchUser & { id: string }

type MentionListProps = {
  items: MentionItem[]
  command: (item: MentionItem) => void
  query: string
}

function MentionList({ items, command, query }: MentionListProps) {
  const [selectedIndex, setSelectedIndex] = React.useState(0)
  const itemRefs = React.useRef<Array<HTMLDivElement | null>>([])

  React.useEffect(() => {
    setSelectedIndex(0)
  }, [items, query])

  React.useEffect(() => {
    itemRefs.current[selectedIndex]?.scrollIntoView({ block: "nearest" })
  }, [selectedIndex])

  if (!items.length) {
    return (
      <div className="mention-list-popup">
        <div className="mention-no-results">未找到匹配用户</div>
      </div>
    )
  }

  return (
    <div className="mention-list-popup">
      {items.map((item, index) => (
        <div
          key={item.id}
          ref={(el) => { itemRefs.current[index] = el }}
          className={`mention-item ${index === selectedIndex ? "is-selected" : ""}`}
          onMouseEnter={() => setSelectedIndex(index)}
          onClick={() => command(item)}
        >
          <img
            src={item.user?.avatar || item.user?.smallAvatar || "/default-avatar.png"}
            alt=""
            className="mention-avatar"
            width={24}
            height={24}
          />
          <div className="mention-info">
            <span className="mention-nickname">{item.nickname || item.user?.nickname || item.username || ""}</span>
            {item.username ? <span className="mention-username">@{item.username}</span> : null}
          </div>
        </div>
      ))}
    </div>
  )
}

function createMentionSuggestion() {
  return Mention.configure({
    HTMLAttributes: {
      class: "inline-mention",
    },
    suggestion: {
      char: "@",
      pluginKey: new PluginKey("mention"),
      items: async ({ query }: { query: string }) => {
        if (!query || query.length < 1) return []
        try {
          const result = await searchUsers({ keyword: query })
          const users = (result?.results || []) as SearchUser[]
          return users.slice(0, 8).map((u) => ({
            ...u,
            id: u.user?.id || "",
          })) as MentionItem[]
        } catch {
          return []
        }
      },
      render: () => {
        let component: ReactRenderer<unknown, MentionListProps> | null = null

        return {
          onStart: (props: SuggestionProps<MentionItem>) => {
            component = new ReactRenderer(MentionList, {
              editor: props.editor,
              props: {
                items: props.items,
                command: (item: MentionItem) => {
                  props.command({ id: item.id, label: `@${item.nickname || item.user?.nickname || item.username || ""}` })
                },
                query: props.query,
              },
            })
            const portalTarget = document.fullscreenElement || document.body
            portalTarget.appendChild(component.element)
            updateMentionPosition(component.element, props.clientRect?.() || null)
          },
          onUpdate: (props: SuggestionProps<MentionItem>) => {
            if (!component) return
            component.updateProps({
              items: props.items,
              command: (item: MentionItem) => {
                props.command({ id: item.id, label: `@${item.nickname || item.user?.nickname || item.username || ""}` })
              },
              query: props.query,
            })
            updateMentionPosition(component.element, props.clientRect?.() || null)
          },
          onKeyDown: (props: SuggestionKeyDownProps) => {
            if (props.event.key === "Escape") {
              component?.destroy()
              component?.element.remove()
              component = null
              return true
            }
            return false
          },
          onExit: () => {
            component?.destroy()
            component?.element.remove()
            component = null
          },
        }
      },
    },
  })
}

function updateMentionPosition(element: HTMLElement, clientRect: (() => DOMRect | null) | DOMRect | null) {
  const rect = typeof clientRect === "function" ? clientRect() : clientRect
  if (!rect) return
  const offset = 6
  let top = rect.bottom + offset
  let left = rect.left
  if (top + 200 > window.innerHeight) {
    top = rect.top - 200 - offset
  }
  if (left + 240 > window.innerWidth) {
    left = window.innerWidth - 240 - 8
  }
  if (left < 8) left = 8
  if (top < 8) top = 8
  element.style.position = "fixed"
  element.style.zIndex = "9999"
  element.style.top = `${top}px`
  element.style.left = `${left}px`
}

// ---- EmojiPicker component ----
function EmojiPicker({
  onSelect,
  onClose,
  position,
}: {
  onSelect: (emoji: string) => void
  onClose: () => void
  position: { top: number; left: number }
}) {
  const [search, setSearch] = React.useState("")
  const filtered = React.useMemo(() => {
    if (!search.trim()) return EMOJI_LIST
    const q = search.toLowerCase()
    return EMOJI_LIST.filter(
      (e) =>
        e.emoji.includes(q) ||
        e.keywords.some((k) => k.includes(q))
    )
  }, [search])

  return (
    <div
      className="emoji-picker-popup"
      style={{ top: position.top, left: position.left }}
    >
      <div className="emoji-search">
        <input
          type="text"
          placeholder="搜索 Emoji…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          autoFocus
        />
      </div>
      <div className="emoji-grid">
        {filtered.map((e) => (
          <button
            key={e.emoji}
            type="button"
            className="emoji-item"
            title={e.keywords[0]}
            onMouseDown={(ev) => ev.preventDefault()}
            onClick={() => {
              onSelect(e.emoji)
              onClose()
            }}
          >
            {e.emoji}
          </button>
        ))}
      </div>
    </div>
  )
}

function createSlashSuggestion(labels: RichTextEditorLabels, locale: string) {
  return Extension.create({
    name: "slash-commands",

    addProseMirrorPlugins() {
      const editor = this.editor
      return [
        Suggestion({
          editor,
          char: "/",
          command: ({ editor, range, props }: { editor: Editor; range: Range; props: SlashCommandItem }) => {
            props.command({ editor, range })
          },
          items: ({ query }: { query: string }) => {
            const searchQuery = query.toLowerCase().trim()
            return slashItems(labels, locale)
              .filter((item) => {
                if (!searchQuery) return true
                return item.title.toLowerCase().includes(searchQuery) || item.description.toLowerCase().includes(searchQuery) || item.aliases.some((alias) => alias.toLowerCase().includes(searchQuery))
              })
              .slice(0, 10)
          },
          render: () => {
            let renderer: ReactRenderer<SlashCommandMenuHandle, SlashCommandMenuProps> | null = null
            let currentProps: SuggestionProps<SlashCommandItem, SlashCommandItem> | null = null

            function syncPosition() {
              if (renderer && currentProps) {
                updateSlashMenuPosition(renderer.element, currentProps.clientRect)
              }
            }

            function mount(props: SuggestionProps<SlashCommandItem, SlashCommandItem>) {
              currentProps = props
              renderer = new ReactRenderer(SlashCommandMenu, {
                editor: props.editor,
                props: { ...props, labels },
              })
              renderer.element.classList.add("slash-commands-popup")
              const portalTarget = document.fullscreenElement || document.body
              portalTarget.appendChild(renderer.element)
              syncPosition()
              window.addEventListener("resize", syncPosition)
              window.addEventListener("scroll", syncPosition, true)
            }

            function cleanup() {
              window.removeEventListener("resize", syncPosition)
              window.removeEventListener("scroll", syncPosition, true)
              renderer?.destroy()
              renderer?.element.remove()
              renderer = null
              currentProps = null
            }

            function update(props: SuggestionProps<SlashCommandItem, SlashCommandItem>) {
              currentProps = props
              renderer?.updateProps({ ...props, labels })
              syncPosition()
            }

            return {
              onStart(props) {
                mount(props)
              },
              onUpdate(props) {
                update(props)
              },
              onKeyDown({ event, view }: SuggestionKeyDownProps) {
                if (event.key === "Escape") {
                  exitSuggestion(view)
                  cleanup()
                  return true
                }
                return renderer?.ref?.onKeyDown(event) || false
              },
              onExit() {
                cleanup()
              },
            }
          },
        }),
      ]
    },
  })
}

function ToolbarButton({
  title,
  active,
  disabled,
  children,
  onClick,
}: {
  title: string
  active?: boolean
  disabled?: boolean
  children: React.ReactNode
  onClick: () => void
}) {
  return (
    <button type="button" className={cn("m-editor-toolbar-button", active && "is-active")} title={title} disabled={disabled} onClick={onClick}>
      {children}
    </button>
  )
}

function ToolbarDivider() {
  return <span className="m-editor-toolbar-divider" />
}

function ColorButton({
  title,
  clearTitle,
  type,
  palette,
  activeColor,
  onApply,
  onClear,
  children,
}: {
  title: string
  clearTitle: string
  type: "text" | "background"
  palette: string[]
  activeColor: string
  onApply: (color: string) => void
  onClear: () => void
  children: React.ReactNode
}) {
  const [open, setOpen] = React.useState(false)
  const buttonRef = React.useRef<HTMLDivElement>(null)
  const popupRef = React.useRef<HTMLDivElement>(null)
  const [popupPosition, setPopupPosition] = React.useState({ top: 0, left: 0 })
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  const updatePopupPosition = React.useCallback(() => {
    const button = buttonRef.current
    if (!button) {
      return
    }
    const rect = button.getBoundingClientRect()
    const popupWidth = popupRef.current?.offsetWidth || 220
    const left = Math.min(Math.max(rect.left, 8), window.innerWidth - popupWidth - 8)
    setPopupPosition({ top: rect.bottom + 6, left })
  }, [])

  React.useEffect(() => {
    if (!open) {
      return
    }
    updatePopupPosition()
    window.addEventListener("resize", updatePopupPosition)
    window.addEventListener("scroll", updatePopupPosition, true)
    return () => {
      window.removeEventListener("resize", updatePopupPosition)
      window.removeEventListener("scroll", updatePopupPosition, true)
    }
  }, [open, updatePopupPosition])

  React.useEffect(() => {
    if (!open) {
      return
    }
    const closeOnOutsideEvent = (event: PointerEvent | FocusEvent) => {
      const target = event.target
      if (!(target instanceof globalThis.Node)) {
        return
      }
      if (buttonRef.current?.contains(target) || popupRef.current?.contains(target)) {
        return
      }
      setOpen(false)
    }
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false)
      }
    }
    document.addEventListener("pointerdown", closeOnOutsideEvent, true)
    document.addEventListener("focusin", closeOnOutsideEvent, true)
    document.addEventListener("keydown", closeOnEscape, true)
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsideEvent, true)
      document.removeEventListener("focusin", closeOnOutsideEvent, true)
      document.removeEventListener("keydown", closeOnEscape, true)
    }
  }, [open])

  const portalTarget = mounted ? document.fullscreenElement || document.body : null
  const popup =
    open && portalTarget
      ? createPortal(
          <div ref={popupRef} className="color-popup" style={{ top: popupPosition.top, left: popupPosition.left }}>
            <div className="color-picker-header">
              <span>{title}</span>
              <button
                type="button"
                className="clear-color"
                title={clearTitle}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  setOpen(false)
                  onClear()
                }}
              >
                <div className="default-color">{!activeColor ? <Check size={16} /> : null}</div>
              </button>
            </div>
            <div className="color-grid">
              {palette.map((color) => (
                <button
                  key={color}
                  type="button"
                  className="color-option"
                  style={{ backgroundColor: color }}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => {
                    setOpen(false)
                    onApply(color)
                  }}
                >
                  {activeColor === color ? <Check size={16} className="check-icon" /> : null}
                </button>
              ))}
            </div>
          </div>,
          portalTarget
        )
      : null

  return (
    <div ref={buttonRef} className={`${type}-color-button m-editor-color-button`}>
      <ToolbarButton title={title} active={Boolean(activeColor)} onClick={() => setOpen((value) => !value)}>
        <span className="button-content">{children}</span>
      </ToolbarButton>
      <div className="color-indicator" style={{ backgroundColor: activeColor || "transparent" }} />
      {popup}
    </div>
  )
}

export function RichTextEditor({
  value,
  height = "400px",
  onChange,
}: {
  value: string
  placeholder?: string
  height?: string
  onChange: (value: string) => void
}) {
  const { locale, t } = useI18n()
  const labels = React.useMemo(() => createEditorLabels(t), [t])
  const { catchError } = useToastActions()
  const containerRef = React.useRef<HTMLDivElement>(null)
  const fileInputRef = React.useRef<HTMLInputElement>(null)
  const lastExternalValueRef = React.useRef(value)
  const [uploading, setUploading] = React.useState(false)
  const [isFullscreen, setIsFullscreen] = React.useState(false)
  const [linkOpen, setLinkOpen] = React.useState(false)
  const [linkText, setLinkText] = React.useState("")
  const [linkUrl, setLinkUrl] = React.useState("")
  const [emojiOpen, setEmojiOpen] = React.useState(false)
  const [emojiPosition, setEmojiPosition] = React.useState({ top: 0, left: 0 })
  const emojiButtonRef = React.useRef<HTMLDivElement>(null)

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        link: false,
        horizontalRule: false,
        codeBlock: false,
      }),
      CodeBlockLowlight.configure({
        lowlight,
        defaultLanguage: null,
      }),
      Link.configure({
        openOnClick: false,
        autolink: true,
        linkOnPaste: true,
        HTMLAttributes: {
          target: "_blank",
          rel: "noopener noreferrer",
        },
      }),
      ResizableImage,
      Underline,
      TextAlign.configure({
        types: ["heading", "paragraph"],
      }),
      TextStyle,
      Color,
      BackgroundColor,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Typography,
      HorizontalRule,
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableCell,
      TableHeader,
      Highlight.configure({
        multicolor: true,
      }),
      FontSize,
      Superscript,
      Subscript,
      Callout,
      createMentionSuggestion(),
      createSlashSuggestion(labels, locale),
      Placeholder.configure({
        placeholder: labels.placeholder,
      }),
    ],
    content: value || "",
    editorProps: {
      attributes: {
        class: "tiptap",
      },
      handlePaste(view, event) {
        const items = event.clipboardData?.items
        if (!items?.length) {
          return false
        }
        const files = Array.from(items)
          .filter((item) => item.type.includes("image"))
          .map((item) => item.getAsFile())
          .filter(Boolean) as File[]
        if (!files.length) {
          return false
        }
        event.preventDefault()
        void uploadImages(files)
        return true
      },
      handleDrop(view, event) {
        const files = Array.from(event.dataTransfer?.files || []).filter((file) => file.type.includes("image"))
        if (!files.length) {
          return false
        }
        event.preventDefault()
        void uploadImages(files)
        return true
      },
    },
    onUpdate({ editor: currentEditor }) {
      const html = currentEditor.getHTML()
      lastExternalValueRef.current = html
      onChange(html === "<p></p>" ? "" : html)
    },
  })

  async function uploadImages(files: File[]) {
    setUploading(true)
    try {
      const urls = await Promise.all(files.map((file) => uploadEditorImage(file)))
      urls.forEach((url, index) => {
        editor?.chain().focus().setResizableImage({ src: url, alt: files[index]?.name || "", title: files[index]?.name || "" }).run()
      })
    } catch (error) {
      catchError(error)
    } finally {
      setUploading(false)
    }
  }

  React.useEffect(() => {
    if (!editor || value === lastExternalValueRef.current) {
      return
    }
    lastExternalValueRef.current = value
    editor.commands.setContent(value || "", { emitUpdate: false })
  }, [editor, value])

  React.useEffect(() => {
    const onFullscreenChange = () => setIsFullscreen(Boolean(document.fullscreenElement))
    document.addEventListener("fullscreenchange", onFullscreenChange)
    return () => document.removeEventListener("fullscreenchange", onFullscreenChange)
  }, [])

  React.useEffect(() => {
    if (!emojiOpen) return
    const button = emojiButtonRef.current
    if (!button) return
    function updatePos() {
      const rect = button?.getBoundingClientRect()
      if (!rect) return
      setEmojiPosition({
        top: Math.min(rect.bottom + 6, window.innerHeight - 330),
        left: Math.min(Math.max(rect.left, 8), window.innerWidth - 288),
      })
    }
    updatePos()
    window.addEventListener("resize", updatePos)
    window.addEventListener("scroll", updatePos, true)
    return () => {
      window.removeEventListener("resize", updatePos)
      window.removeEventListener("scroll", updatePos, true)
    }
  }, [emojiOpen])

  React.useEffect(() => {
    if (!emojiOpen) return
    function onPointerDown(e: PointerEvent) {
      const target = e.target as globalThis.Node | null
      if (!target) return
      if (emojiButtonRef.current?.contains(target)) return
      const popup = document.querySelector(".emoji-picker-popup")
      if (popup?.contains(target)) return
      setEmojiOpen(false)
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setEmojiOpen(false)
    }
    document.addEventListener("pointerdown", onPointerDown, true)
    document.addEventListener("keydown", onKeyDown, true)
    return () => {
      document.removeEventListener("pointerdown", onPointerDown, true)
      document.removeEventListener("keydown", onKeyDown, true)
    }
  }, [emojiOpen])

  function toggleFullscreen() {
    const el = containerRef.current
    if (!el) return
    if (!document.fullscreenElement) {
      void el.requestFullscreen()
    } else {
      void document.exitFullscreen()
    }
  }

  function openLinkDialog() {
    if (!editor) return
    const previousUrl = editor.getAttributes("link").href as string | undefined
    const selectedText = editor.state.doc.textBetween(editor.state.selection.from, editor.state.selection.to, " ")
    setLinkText(selectedText)
    setLinkUrl(previousUrl || "")
    setLinkOpen(true)
  }

  function applyLink() {
    if (!editor) return
    if (!linkUrl) {
      editor.chain().focus().extendMarkRange("link").unsetLink().run()
      setLinkOpen(false)
      return
    }
    if (linkText && editor.state.selection.empty) {
      editor.chain().focus().insertContent(`<a href="${linkUrl}" target="_blank" rel="noopener noreferrer">${linkText}</a>`).run()
    } else {
      editor.chain().focus().extendMarkRange("link").setLink({ href: linkUrl }).run()
    }
    setLinkOpen(false)
  }

  const activeTextColor = (editor?.getAttributes("textStyle").color as string | undefined) || ""
  const activeBackgroundColor = (editor?.getAttributes("textStyle").backgroundColor as string | undefined) || ""
  const toolbar = labels.toolbar

  return (
    <div ref={containerRef} className="m-editor-container" style={{ height }}>
      <div className="editor-toolbar">
        <div className="editor-toolbar-btns editor-toolbar-left">
          <ToolbarButton title={toolbar.undo} onClick={() => editor?.chain().focus().undo().run()}>
            <Undo2 size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.redo} onClick={() => editor?.chain().focus().redo().run()}>
            <Redo2 size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.bold} active={editor?.isActive("bold")} onClick={() => editor?.chain().focus().toggleBold().run()}>
            <Bold size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.underline} active={editor?.isActive("underline")} onClick={() => editor?.chain().focus().toggleUnderline().run()}>
            <UnderlineIcon size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.italic} active={editor?.isActive("italic")} onClick={() => editor?.chain().focus().toggleItalic().run()}>
            <Italic size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.strike} active={editor?.isActive("strike")} onClick={() => editor?.chain().focus().toggleStrike().run()}>
            <Strikethrough size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.superscript} active={editor?.isActive("superscript")} onClick={() => editor?.chain().focus().toggleSuperscript().run()}>
            <SuperscriptIcon size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.subscript} active={editor?.isActive("subscript")} onClick={() => editor?.chain().focus().toggleSubscript().run()}>
            <SubscriptIcon size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.heading1} active={editor?.isActive("heading", { level: 1 })} onClick={() => editor?.chain().focus().toggleHeading({ level: 1 }).run()}>
            <Heading1 size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.heading2} active={editor?.isActive("heading", { level: 2 })} onClick={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()}>
            <Heading2 size={16} />
          </ToolbarButton>
          <select
            className="m-editor-font-size-select"
            title={toolbar.fontSize}
            value={editor?.getAttributes("textStyle").fontSize || ""}
            onChange={(event) => {
              const value = event.target.value
              if (value) {
                editor?.chain().focus().setFontSize(value).run()
              } else {
                editor?.chain().focus().unsetFontSize().run()
              }
            }}
          >
            <option value="">{toolbar.fontSize}</option>
            {FONT_SIZES.map((s) => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.quote} active={editor?.isActive("blockquote")} onClick={() => editor?.chain().focus().toggleBlockquote().run()}>
            <Quote size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.bulletList} active={editor?.isActive("bulletList")} onClick={() => editor?.chain().focus().toggleBulletList().run()}>
            <List size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.orderedList} active={editor?.isActive("orderedList")} onClick={() => editor?.chain().focus().toggleOrderedList().run()}>
            <ListOrdered size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.taskList} active={editor?.isActive("taskList")} onClick={() => editor?.chain().focus().toggleTaskList().run()}>
            <ListTodo size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.alignLeft} active={editor?.isActive({ textAlign: "left" })} onClick={() => editor?.chain().focus().setTextAlign("left").run()}>
            <AlignLeft size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.alignCenter} active={editor?.isActive({ textAlign: "center" })} onClick={() => editor?.chain().focus().setTextAlign("center").run()}>
            <AlignCenter size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.alignRight} active={editor?.isActive({ textAlign: "right" })} onClick={() => editor?.chain().focus().setTextAlign("right").run()}>
            <AlignRight size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ColorButton title={toolbar.textColor} clearTitle={toolbar.clearColor} type="text" palette={TEXT_COLOR_PALETTE} activeColor={activeTextColor} onApply={(color) => editor?.chain().focus().setColor(color).run()} onClear={() => editor?.chain().focus().unsetColor().run()}>
            <Palette size={16} />
          </ColorButton>
          <ColorButton title={toolbar.backgroundColor} clearTitle={toolbar.clearColor} type="background" palette={BACKGROUND_COLOR_PALETTE} activeColor={activeBackgroundColor} onApply={(color) => editor?.chain().focus().setBackgroundColor(color).run()} onClear={() => editor?.chain().focus().unsetBackgroundColor().run()}>
            <Paintbrush size={16} />
          </ColorButton>
          <ToolbarDivider />
          <ToolbarButton title={toolbar.inlineCode} active={editor?.isActive("code")} onClick={() => editor?.chain().focus().toggleCode().run()}>
            <Code size={16} />
          </ToolbarButton>
          <ToolbarButton title={toolbar.codeBlock} active={editor?.isActive("codeBlock")} onClick={() => editor?.chain().focus().toggleCodeBlock().run()}>
            <Code2 size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton
            title="信息 Callout"
            onClick={() => editor?.chain().focus().setCallout({ type: "info" }).run()}
          >
            ℹ️
          </ToolbarButton>
          <ToolbarButton
            title="警告 Callout"
            onClick={() => editor?.chain().focus().setCallout({ type: "warning" }).run()}
          >
            ⚠️
          </ToolbarButton>
          <ToolbarButton
            title="提示 Callout"
            onClick={() => editor?.chain().focus().setCallout({ type: "tip" }).run()}
          >
            💡
          </ToolbarButton>
          <ToolbarButton
            title="成功 Callout"
            onClick={() => editor?.chain().focus().setCallout({ type: "success" }).run()}
          >
            ✅
          </ToolbarButton>
          <ToolbarDivider />
          <ToolbarButton
            title={toolbar.table}
            onClick={() =>
              editor
                ?.chain()
                .focus()
                .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
                .run()
            }
          >
            <TableIcon size={16} />
          </ToolbarButton>
          <ToolbarButton
            title={toolbar.highlight}
            active={editor?.isActive("highlight")}
            onClick={() => editor?.chain().focus().toggleHighlight().run()}
          >
            <Highlighter size={16} />
          </ToolbarButton>
          <div ref={emojiButtonRef} className="emoji-picker-wrapper">
            <ToolbarButton
              title="Emoji"
              active={emojiOpen}
              onClick={() => setEmojiOpen((v) => !v)}
            >
              <Smile size={16} />
            </ToolbarButton>
            {emojiOpen ? (
              <EmojiPicker
                position={emojiPosition}
                onSelect={(emoji) => {
                  editor?.chain().focus().insertContent(emoji).run()
                }}
                onClose={() => setEmojiOpen(false)}
              />
            ) : null}
          </div>
          <ToolbarButton
            title={toolbar.clearFormat}
            onClick={() => editor?.chain().focus().clearNodes().unsetAllMarks().run()}
          >
            <Eraser size={16} />
          </ToolbarButton>
          <ToolbarDivider />
          {/* ---- Table contextual actions ---- */}
          {editor?.isActive("table") ? (
            <>
              <ToolbarButton title={toolbar.addRowBefore} onClick={() => editor?.chain().focus().addRowBefore().run()}>
                <Plus size={14} className="table-action-icon" /><ArrowLeft size={14} />
              </ToolbarButton>
              <ToolbarButton title={toolbar.addRowAfter} onClick={() => editor?.chain().focus().addRowAfter().run()}>
                <Plus size={14} className="table-action-icon" /><ArrowRight size={14} />
              </ToolbarButton>
              <ToolbarButton title={toolbar.addColBefore} onClick={() => editor?.chain().focus().addColumnBefore().run()}>
                <ArrowLeft size={14} /><Plus size={14} className="table-action-icon" />
              </ToolbarButton>
              <ToolbarButton title={toolbar.addColAfter} onClick={() => editor?.chain().focus().addColumnAfter().run()}>
                <ArrowRight size={14} /><Plus size={14} className="table-action-icon" />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton title={toolbar.deleteRow} onClick={() => editor?.chain().focus().deleteRow().run()}>
                <Trash2 size={14} /><span className="table-action-label">↕</span>
              </ToolbarButton>
              <ToolbarButton title={toolbar.deleteCol} onClick={() => editor?.chain().focus().deleteColumn().run()}>
                <Trash2 size={14} /><span className="table-action-label">↔</span>
              </ToolbarButton>
              <ToolbarButton title={toolbar.deleteTable} onClick={() => editor?.chain().focus().deleteTable().run()}>
                <Trash2 size={14} /><Table2 size={14} />
              </ToolbarButton>
              <ToolbarDivider />
            </>
          ) : null}
          <ToolbarDivider />
          <div className="link-button">
            <ToolbarButton title={toolbar.link} active={editor?.isActive("link")} onClick={openLinkDialog}>
              <LinkIcon size={16} />
            </ToolbarButton>
            {linkOpen ? (
              <div className="link-dialog-content">
                <div className="link-input-group">
                  <label>{labels.linkDialog.textLabel}</label>
                  <input value={linkText} placeholder={labels.linkDialog.textPlaceholder} onChange={(event) => setLinkText(event.currentTarget.value)} />
                </div>
                <div className="link-input-group">
                  <label>{labels.linkDialog.urlLabel}</label>
                  <input value={linkUrl} placeholder={labels.linkDialog.urlPlaceholder} onChange={(event) => setLinkUrl(event.currentTarget.value)} />
                </div>
                <div className="link-dialog-actions">
                  <button type="button" className="btn-primary" onClick={applyLink}>{labels.linkDialog.confirm}</button>
                  <button type="button" className="btn-danger" onClick={() => { editor?.chain().focus().extendMarkRange("link").unsetLink().run(); setLinkOpen(false) }}>{labels.linkDialog.remove}</button>
                  <button type="button" className="btn-secondary" onClick={() => setLinkOpen(false)}>{labels.linkDialog.cancel}</button>
                </div>
              </div>
            ) : null}
          </div>
          <div className="image-upload-button">
            <ToolbarButton title={toolbar.image} disabled={!editor || uploading} onClick={() => fileInputRef.current?.click()}>
              <ImageIcon size={16} />
            </ToolbarButton>
            {uploading ? (
              <div className="upload-progress">
                <div className="upload-spinner" />
                <span>{toolbar.uploading}</span>
              </div>
            ) : null}
          </div>
          <ToolbarButton title={toolbar.horizontalRule} onClick={() => editor?.chain().focus().setHorizontalRule().run()}>
            <MinusSquare size={16} />
          </ToolbarButton>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(event) => {
              const files = Array.from(event.currentTarget.files || [])
              if (files.length) void uploadImages(files)
              event.currentTarget.value = ""
            }}
          />
        </div>
        <div className="editor-toolbar-btns editor-toolbar-right">
          <ToolbarButton title={isFullscreen ? toolbar.exitFullscreen : toolbar.fullscreen} active={isFullscreen} onClick={toggleFullscreen}>
            {isFullscreen ? <Minimize size={16} /> : <Maximize size={16} />}
          </ToolbarButton>
        </div>
      </div>
      <EditorContent editor={editor} className="editor-content" />
    </div>
  )
}
