import React from 'react';
import { Icon } from '@material-ui/core';
import '../css/Feature.scss';

export default function Feature(props) {
  const { icon, feature, description } = props;
  return (
    <div className="feature">
      <div className="featureHeader">
        <Icon className="featureIcon">{icon}</Icon>
        <h3>{feature}</h3>
      </div>
      <p>{description}</p>
    </div>
  );
}
