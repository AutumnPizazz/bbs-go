import type * as React from "react"
import { Link as RouterLink } from "react-router-dom"

type LinkProps = Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href"> & {
  href: string
  prefetch?: boolean
}

export default function Link({ href, prefetch: _prefetch, ...props }: LinkProps) {
  if (/^(https?:)?\/\//.test(href)) {
    return <a href={href} {...props} />
  }

  return <RouterLink to={href} {...props} />
}
