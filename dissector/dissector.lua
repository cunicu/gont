-- SPDX-FileCopyrightText: 2014-2015 Kim Alvefur
-- SPDX-License-Identifier: MIT

cbor = require('cbor')

trace = Proto('trace', 'Gont Tracing Protocol')

-- Protocol fields
trace.fields.msg       = ProtoField.string('trace.message',     'Message')
trace.fields.type      = ProtoField.string('trace.type',        'Type')
trace.fields.source    = ProtoField.string('trace.source',      'Source')
trace.fields.level     = ProtoField.string('trace.level',       'Level')
trace.fields.level_num = ProtoField.int32('trace.level_num',    'Level number', base.DEC)
trace.fields.pid       = ProtoField.uint32('trace.pid',         'Process ID',   base.DEC)
trace.fields.func      = ProtoField.string('trace.function',    'Function')
trace.fields.file      = ProtoField.string('trace.file',        'Filename')
trace.fields.line      = ProtoField.uint32('trace.line',        'Line',         base.DEC)
trace.fields.data      = ProtoField.string('trace.data',        'Data')

-- Zap log level names
level_names = {
    "debug",
    "info",
    "warn",
    "error",
    "dpanic",
    "panic",
    "fatal"
}

function build_tree(l, o, t)
        local st = t:add(l)
        for k, v in pairs(o) do
            if type(v) == 'table' then
                build_tree(k, v, st)
            else
                st:add(string.format('%s:', k), v)
            end
       end
end

function trace.dissector(buffer, pinfo, tree)
    dec = cbor.decode(buffer:raw())

    pinfo.cols.protocol = "Gont"

    local subtree = tree:add(trace)

    if dec.type ~= nil then
        pinfo.cols.protocol:append(" "..dec.type)
    end
    if dec.msg ~= nil then
        pinfo.cols.info = dec.msg
        subtree:add(trace.fields.msg,  dec.msg)
    end
    if dec.src ~= nil then
        pinfo.cols.src = dec.src
        subtree:add(trace.fields.source, dec.src)
    end
    if dec.type ~= nil then
        subtree:add(trace.fields.type, dec.type)
    end
    if dec.lvl ~= nil and dec.type == "log" then
        local lvl = level_names[dec.lvl]
        pinfo.cols.protocol:append("("..lvl..")")
        subtree:add(trace.fields.level_num, dec.lvl)
        subtree:add(trace.fields.level, lvl)
    end
    if dec.pid ~= nil then
        pinfo.cols.dst = dec.pid
        subtree:add(trace.fields.pid,  dec.pid)
    end
    if dec.func ~= nil then
        subtree:add(trace.fields.func, dec.func)
    end
    if dec.file ~= nil then
        subtree:add(trace.fields.file, dec.file)
    end
    if dec.line ~= nil then
        subtree:add(trace.fields.line, dec.line)
    end
    if dec.data ~= nil then
        if type(dec.data) == "table" then
            build_tree(trace.fields.data, dec.data, subtree)
        else
            subtree:add(trace.fields.data, dec.data)
        end
    end
end

local udlt = DissectorTable.get('wtap_encap')
udlt:add(wtap.USER0, trace)
