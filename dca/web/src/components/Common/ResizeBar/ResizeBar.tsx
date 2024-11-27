export default function ResizeBar({oriWidth, setWidth, className, maxWidth = 500, minWidth = 150}) {
  let startX = 0
  let width = oriWidth
  const resize = (e) => {
    startX = e.pageX
    const updateWidth = (e) => {
      const deltaX = e.pageX - startX
      width = width + deltaX
      if (width <= minWidth) {
        width = minWidth
      } else if (width > maxWidth) {
        width = maxWidth
      } else {
        startX = e.pageX
      }
      setWidth(width)
    }
    document.onmousemove = (e) => {
      updateWidth(e)
    }
    document.onmouseup = () => {
      document.onmousemove = null
      document.onmouseup = null
    }
    return e.preventDefault()
  }
  return <div className={className} onMouseDown={resize}></div>
}