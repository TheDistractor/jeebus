ng = angular.module 'myApp'

ng.directive 'circuitEditor', ->
  restrict: 'E'
  
  scope:
    defs: '='
    data: '='
    
  link: (scope, elem, attr) ->
    svg = d3.select(elem[0]).append 'svg'
      .attr height: "70%"

    # pre-calculate sizes and relative pin coordinates
    for n, d of scope.defs
      d.name or= n
      ins = 0
      for p in d.pins
        p.x = d.width / 2
        if p.dir is 'in'
          p.x = -p.x
          ++ins
      outs = d.pins.length - ins
      step = 16
      yIn = - (ins - 1) * step / 2
      yOut = - (outs - 1) * step / 2
      for p in d.pins
        if p.dir is 'in'
          p.y = yIn
          yIn += step
        else
          p.y = yOut
          yOut += step
      d.height = 30 + step * (if ins > outs then ins else outs)

    findPin = (name) ->
      [gid,pname] = name.split '.'
      for g in scope.data.gadgets
        if gid is g.id
          for p in g.def.pins
            if pname is p.name
              # reverses x and y and uses projection to get horizontal splines
              return y: g.x + p.x + .5, x: g.y + p.y + .5, g: g, p: p

    for d in scope.data.gadgets
      d.def = scope.defs[d.type]
      d.hw = d.def.width / 2
      d.hh = d.def.height / 2

    for d in scope.data.wires
      d.source = findPin d.from
      d.target = findPin d.to

    gadgets = svg.selectAll('.gadget').data(scope.data.gadgets)
    wires = svg.selectAll('.wire').data(scope.data.wires)

    diag = d3.svg.diagonal()
      .projection (d) -> [d.y, d.x] # undo the x/y reversal from findPin
    
    gadgetDrag = d3.behavior.drag()
      .origin Object
      .on 'dragstart', (d) ->
        @parentNode.appendChild @ # move to front
      .on 'drag', (d) ->
        d.x = d3.event.x | 0 # stay on int coordinates
        d.y = d3.event.y | 0 # stay on int coordinates
        d3.select(@).attr
          transform: (d) -> "translate(#{d.x},#{d.y})"
        # recalculate endpoints and redraw all wires attached to this gadget
        wires.filter (w) -> w.source.g is d or w.target.g is d
          .each (d) ->
            d.source = findPin d.from
            d.target = findPin d.to
          .attr d: diag
      .on 'dragend', (d) ->
        console.log 'save gadget', d # TODO: save to server

    g = gadgets.enter().append('g').call(gadgetDrag)
      .attr class: 'gadget'
    g.append('rect')
      .each (d) ->
        d3.select(@).attr
          class: 'outline'
          # 1px lines render sharply when on a 0.5px offset
          x: 0.5 - d.hw
          y: 0.5 - d.hh
          width: 2 * d.hw
          height: 2 * d.hh
      .style fill: (d) -> d.def.shade
    g.append('text').text (d) -> d.title
      .attr class: 'title', y: (d) -> 12 - d.hh
    g.append('text').text (d) -> d.def.name
      .attr class: 'type', y: (d) -> -4 + d.hh
    g.append('text').text (d) -> d.def.icon
      .attr class: 'iconfont', x: 0, y: 0
        
    pins = gadgets.selectAll('rect .pin').data (d) -> d.def.pins
    p = pins.enter()
    p.append('circle')
      .attr class: 'pin', cx: ((d) -> d.x+.5), cy: ((d) -> d.y+.5), r: 3
      .on 'mousedown', (d) ->
        console.log 'c1', d
    p.append('text').text (d) -> d.name
      .attr
        class: (d) -> d.dir
        x: (d) -> if d.dir is 'in' then d.x + 7 else d.x - 7
        y: (d) -> d.y + 5

    wires.enter().insert('path', 'g') # uses insert to move to back right away
      .attr class: 'wire', d: diag

    gadgets.attr transform: (d) -> "translate(#{d.x},#{d.y})"
