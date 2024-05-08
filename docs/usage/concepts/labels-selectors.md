# Labels and Selectors Guide

## Intended Audience

This documentation is written for Data Tool users who pull and retrieve bottles; these users are referred to in Data Tool's docs as *consumers*, *researchers*, or *AI/ML practitioners*. These users are often locating bottles in a telemetry server.

> - Consult the [Data Tool User Guide](../user-guide.md) to review Data Tools's key concepts and common usage patterns

## Prerequisites

This guide does not have prerequisites for the intended audience.

## Labels

Labels are the way that bottle authors can classify and identify sets of bottles. Commonly, labels are used to specify the:

- Project that created it, e.g. `project=CivHire`
- Type of bottle, e.g. `type=model`, `problem=classification`
- (Hyper)Parameters used to generate the model, e.g. `learning-rate=0.005`, `layers=7`
- Source code version, e.g. `version=v0.2.4`

In short labels are used to identify anything a user might want to use to select bottles. The selection of bottles based on labels is done with label selectors.

Labels can also be used on parts of bottles. For example a bottle representing a data set might have both the training and test sets contained within it as different parts. In that case you would want to label them set with something like `subset=train` and `subset=test`, respectively. That way a user can restrict what they download if only a few parts are required.

## Label Selectors

Label selectors (or simply selectors) are used to restrict the set of bottles. Assuming users label their bottles appropriately, you can effectively make arbitrary groups of bottles by picking which selectors must match.

The syntax used is the same as Kubernetes [label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors). Eventually this might be extended to become a superset of what Kubernetes provides.

A selector is composed of zero or more requirements. For example, the selector `project=CivHire`,`type=model` has two requirements `project=CivHire` and `type=model` that must match the bottle in order to be selected. All requirements must match for a selector to match. So the `,` in `project=CivHire,type=model` semantically means `AND`.

When multiple multiple selectors are specified, any one of those selectors must match. So semantically selectors are OR’d together. In mathematical terms this is disjunctive normal form (e.g., (r₀ ∧ r₁ ∧ …) ∨ (r₃ ∧ …)). In electrical engineering terms, this is the *sum of products* form (i.e., r₀ r₁ + r₃).

There many ways to specify a requirement. Above we only showed the simplest form `somekey=somevalue`. Note that whitespace within a requirement is insignificant. Here are all the available ways to specify a requirement:

1. `key = value` and `key == value` both require that the value of key equal value
2. `key != value` requires that the value of key not be value
3. `key in (value1, value2)` requires that the value of key be either value1 or value24.
4. `key notin (value1, value2)` requires that the value of key be neither value1 or value2
5. `key` requires that a label with key exist (the value of key is not checked)
6. `!key` requires that a label with key not exist (the value of key is not checked)
7. `key < 5` requires that the value of key be less than 5
8. `key > 7` requires that the value of key be greater than 7

In some situations there is a temporary restriction that inequality requirements (7 and 8) only work when the value is an integer.
