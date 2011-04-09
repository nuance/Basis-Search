Basis Search
============

Component-based search. Hold your text index on one node, your geo
index on another, and your attributes on a third. Queries are executed
on all relevant classes of nodes, and results are streamed back
through an aggregator, which collects query statistics and selects the
top results.

Aggegrators need to have fast access to any attributes not queries
over that you're interested on collecting statistics about.

